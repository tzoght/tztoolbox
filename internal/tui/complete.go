package tui

import (
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// completeLine returns the (possibly extended) line after a Tab key press.
// Strategy:
//
//   - If the user is mid-token at position 0, complete the subcommand name.
//   - If the user is past a known subcommand, complete its long flags.
//   - Otherwise return the line unchanged.
//
// We complete the longest unambiguous prefix; if that doesn't extend the
// token but multiple matches exist, the second value (matches) is non-empty
// so the model can show suggestions.
func completeLine(root *cobra.Command, line string) (newLine string, suggestions []string) {
	tokens := strings.Fields(line)
	endsWithSpace := strings.HasSuffix(line, " ")

	switch {
	case len(tokens) == 0:
		return line, listSubcommands(root, "")

	case len(tokens) == 1 && !endsWithSpace:
		// Completing the first word as a subcommand name.
		matches := listSubcommands(root, tokens[0])
		if len(matches) == 0 {
			return line, nil
		}
		if len(matches) == 1 {
			return matches[0] + " ", nil
		}
		prefix := commonPrefix(matches)
		if len(prefix) > len(tokens[0]) {
			return prefix, matches
		}
		return line, matches

	default:
		// Find the subcommand to complete flags for. For now, only top-level
		// subcommands have completable flags.
		var sub *cobra.Command
		for _, c := range root.Commands() {
			if c.Hidden {
				continue
			}
			if c.Name() == tokens[0] {
				sub = c
				break
			}
		}
		if sub == nil {
			return line, nil
		}
		// Complete the current token if it starts with "--", else nothing.
		var current string
		if !endsWithSpace && len(tokens) > 1 {
			current = tokens[len(tokens)-1]
		}
		if !strings.HasPrefix(current, "--") {
			return line, nil
		}
		flags := listFlags(sub, current)
		if len(flags) == 0 {
			return line, nil
		}
		if len(flags) == 1 {
			tokens[len(tokens)-1] = flags[0]
			return strings.Join(tokens, " ") + " ", nil
		}
		prefix := commonPrefix(flags)
		if len(prefix) > len(current) {
			tokens[len(tokens)-1] = prefix
			return strings.Join(tokens, " "), flags
		}
		return line, flags
	}
}

// listSubcommands returns visible top-level subcommand names matching prefix.
func listSubcommands(root *cobra.Command, prefix string) []string {
	var out []string
	for _, c := range root.Commands() {
		if c.Hidden || c.Name() == "help" || c.Name() == "completion" {
			continue
		}
		if prefix == "" || strings.HasPrefix(c.Name(), prefix) {
			out = append(out, c.Name())
		}
	}
	sort.Strings(out)
	return out
}

// listFlags returns the long form of every visible flag on cmd matching the
// supplied --prefix.
func listFlags(cmd *cobra.Command, prefix string) []string {
	prefix = strings.TrimPrefix(prefix, "--")
	var out []string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			out = append(out, f.Name)
		}
	})
	cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) {
		if !f.Hidden {
			out = append(out, f.Name)
		}
	})
	flags := dedupeAndSort(out)
	matched := flags[:0]
	for _, name := range flags {
		if strings.HasPrefix(name, prefix) {
			matched = append(matched, "--"+name)
		}
	}
	return matched
}

// commonPrefix returns the longest common prefix across xs. Empty input
// returns "".
func commonPrefix(xs []string) string {
	if len(xs) == 0 {
		return ""
	}
	p := xs[0]
	for _, s := range xs[1:] {
		i := 0
		for i < len(p) && i < len(s) && p[i] == s[i] {
			i++
		}
		p = p[:i]
		if p == "" {
			return ""
		}
	}
	return p
}

func dedupeAndSort(xs []string) []string {
	seen := map[string]struct{}{}
	for _, x := range xs {
		seen[x] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for x := range seen {
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}
