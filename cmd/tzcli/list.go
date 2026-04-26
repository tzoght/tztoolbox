package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tzoght/tztoolbox/internal/authoring"
	"github.com/tzoght/tztoolbox/internal/ui"
)

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list [commands|rules|skills]",
		Short: "List artifacts under shared/",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot()
			if err != nil {
				return err
			}
			var kinds []authoring.Kind
			if len(args) == 1 {
				k, err := parseListKind(args[0])
				if err != nil {
					return err
				}
				kinds = []authoring.Kind{k}
			}
			items, err := authoring.List(root, kinds)
			if err != nil {
				return err
			}
			if flagJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(items)
			}
			r := ui.From(cmd.Context())
			rows := make([][]ui.Cell, 0, len(items))
			for _, it := range items {
				rows = append(rows, []ui.Cell{
					ui.PlainCell(string(it.Kind)),
					ui.StyledCell(it.Name, ui.StyleBold),
					ui.PlainCell(fmt.Sprintf("%d", it.Size)),
					ui.StyledCell(it.Path, ui.StylePath),
				})
			}
			r.Table([]string{"KIND", "NAME", "SIZE", "PATH"}, rows)
			return nil
		},
	}
	return cmd
}

// parseListKind accepts the plural forms used in the help text.
func parseListKind(s string) (authoring.Kind, error) {
	switch s {
	case "command", "commands":
		return authoring.KindCommand, nil
	case "rule", "rules":
		return authoring.KindRule, nil
	case "skill", "skills":
		return authoring.KindSkill, nil
	}
	return "", errorf("unknown kind %q (expected commands|rules|skills)", s)
}
