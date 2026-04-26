package tui

import (
	"bufio"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// historyMax bounds how many lines we keep on disk to avoid unbounded growth.
const historyMax = 1000

// historyFilePath returns the path to the persistent history file. It uses
// ${XDG_CONFIG_HOME}/tzcli/history when set, else ~/.config/tzcli/history.
// On error (no home dir) it returns "" so callers can disable persistence.
func historyFilePath() string {
	if d := os.Getenv("TZCLI_HISTORY_FILE"); d != "" {
		return d
	}
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "tzcli", "history")
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, ".config", "tzcli", "history")
}

// loadHistory reads the history file into memory. Missing file -> empty
// slice (not an error). Any I/O error returns it; callers may choose to
// proceed with an empty history.
func loadHistory(path string) ([]string, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path) //nolint:gosec // user-owned file under config dir
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer f.Close() //nolint:errcheck

	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := strings.TrimRight(sc.Text(), "\r")
		if line != "" {
			out = append(out, line)
		}
	}
	if err := sc.Err(); err != nil {
		return out, err
	}
	if len(out) > historyMax {
		out = out[len(out)-historyMax:]
	}
	return out, nil
}

// saveHistory atomically replaces the history file. Best-effort; errors are
// logged at the call site, never propagated to the model. Empty path skips.
func saveHistory(path string, lines []string) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if len(lines) > historyMax {
		lines = lines[len(lines)-historyMax:]
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".history-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	for _, l := range lines {
		if _, err := tmp.WriteString(l + "\n"); err != nil {
			_ = tmp.Close()
			_ = os.Remove(tmpName)
			return err
		}
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	return os.Rename(tmpName, path)
}

// pushHistory appends line if it differs from the last entry, deduping
// consecutive repeats. Returns the (possibly trimmed) new history.
func pushHistory(h []string, line string) []string {
	line = strings.TrimSpace(line)
	if line == "" {
		return h
	}
	if len(h) > 0 && h[len(h)-1] == line {
		return h
	}
	h = append(h, line)
	if len(h) > historyMax {
		h = h[len(h)-historyMax:]
	}
	return h
}
