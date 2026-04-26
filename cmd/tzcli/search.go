package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tzoght/tztoolbox/internal/diffsearch"
	"github.com/tzoght/tztoolbox/internal/ui"
)

func newSearchCmd() *cobra.Command {
	var inFlag string
	var ignoreCase bool
	cmd := &cobra.Command{
		Use:   "search <pattern>",
		Short: "Regex search across shared/ content",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot()
			if err != nil {
				return err
			}
			var kinds []string
			if inFlag != "" {
				for _, k := range strings.Split(inFlag, ",") {
					k = strings.TrimSpace(k)
					switch k {
					case "command", "commands":
						kinds = append(kinds, "commands")
					case "rule", "rules":
						kinds = append(kinds, "rules")
					case "skill", "skills":
						kinds = append(kinds, "skills")
					default:
						return errorf("unknown kind %q (expected commands|rules|skills)", k)
					}
				}
			}
			hits, err := diffsearch.Search(root, args[0], diffsearch.SearchOptions{Kinds: kinds, IgnoreCase: ignoreCase})
			if err != nil {
				return err
			}
			if flagJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(hits)
			}
			r := ui.From(cmd.Context())
			if len(hits) == 0 {
				r.Note("no matches")
				return nil
			}
			rows := make([][]ui.Cell, 0, len(hits))
			for _, h := range hits {
				rows = append(rows, []ui.Cell{
					ui.StyledCell(h.Path, ui.StylePath),
					ui.PlainCell(fmt.Sprintf("%d", h.Line)),
					ui.PlainCell(h.Text),
				})
			}
			r.Table([]string{"PATH", "LINE", "MATCH"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&inFlag, "in", "", "limit search to specific kinds (comma-separated): commands,rules,skills")
	cmd.Flags().BoolVarP(&ignoreCase, "ignore-case", "i", false, "case-insensitive matching")
	return cmd
}
