package tui

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPushHistoryDedupesConsecutive(t *testing.T) {
	h := []string{"a", "b"}
	h = pushHistory(h, "b")
	if len(h) != 2 {
		t.Fatalf("expected dedup, got %v", h)
	}
	h = pushHistory(h, "c")
	if !reflect.DeepEqual(h, []string{"a", "b", "c"}) {
		t.Fatalf("unexpected history: %v", h)
	}
	h = pushHistory(h, "")
	if !reflect.DeepEqual(h, []string{"a", "b", "c"}) {
		t.Fatalf("empty entry must be ignored: %v", h)
	}
}

func TestPushHistoryRespectsMax(t *testing.T) {
	var h []string
	for i := 0; i < historyMax+50; i++ {
		h = pushHistory(h, "cmd-"+itoa(i))
	}
	if len(h) != historyMax {
		t.Fatalf("expected %d entries, got %d", historyMax, len(h))
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "history")
	want := []string{"sync", "doctor", "installed --tool cursor"}
	if err := saveHistory(path, want); err != nil {
		t.Fatalf("saveHistory: %v", err)
	}
	got, err := loadHistory(path)
	if err != nil {
		t.Fatalf("loadHistory: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round-trip mismatch:\nwant %v\n got %v", want, got)
	}
}

func TestLoadHistoryMissingFileReturnsEmpty(t *testing.T) {
	got, err := loadHistory(filepath.Join(t.TempDir(), "no-such-file"))
	if err != nil {
		t.Fatalf("expected nil error for missing file, got %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestSaveHistoryEmptyPathSkips(t *testing.T) {
	if err := saveHistory("", []string{"x"}); err != nil {
		t.Fatalf("empty path should be a no-op, got %v", err)
	}
}

func TestHistoryFilePathHonorsEnv(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("TZCLI_HISTORY_FILE", filepath.Join(dir, "explicit"))
	if got := historyFilePath(); got != filepath.Join(dir, "explicit") {
		t.Fatalf("expected explicit override, got %q", got)
	}

	t.Setenv("TZCLI_HISTORY_FILE", "")
	t.Setenv("XDG_CONFIG_HOME", dir)
	want := filepath.Join(dir, "tzcli", "history")
	if got := historyFilePath(); got != want {
		t.Fatalf("XDG path mismatch: want %q got %q", want, got)
	}
}

// itoa avoids pulling strconv into this test for a 4-line helper.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := []byte{}
	neg := n < 0
	if neg {
		n = -n
	}
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	if neg {
		buf = append([]byte{'-'}, buf...)
	}
	return string(buf)
}

// keep imports honest if we ever add an os.Stat assertion later.
var _ = os.Stat
