package main

import (
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/tzoght/tztoolbox/internal/authoring"
	"github.com/tzoght/tztoolbox/internal/ui"
)

func newValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Lint the artifact corpus under shared/",
		Long: `Run a battery of consistency checks across shared/:

- SKILL.md frontmatter (name, description ≤200 chars, third-person phrasing)
- command and rule files have a single H1
- no obvious credential tokens are committed`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot()
			if err != nil {
				return err
			}
			issues, err := authoring.Validate(root)
			if err != nil {
				return err
			}
			if flagJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(issues)
			}
			r := ui.From(cmd.Context())
			if len(issues) == 0 {
				r.Note("ok: no issues found")
				return nil
			}
			errors := 0
			rows := make([][]ui.Cell, 0, len(issues))
			for _, i := range issues {
				sevStyle := ui.StyleWarn
				if i.Severity == "error" {
					sevStyle = ui.StyleError
					errors++
				}
				rows = append(rows, []ui.Cell{
					ui.StyledCell(i.Severity, sevStyle),
					ui.StyledCell(i.Path, ui.StylePath),
					ui.PlainCell(i.Message),
				})
			}
			r.Table([]string{"SEVERITY", "PATH", "MESSAGE"}, rows)
			if errors > 0 {
				return errorf("validate found %d error(s) and %d total issue(s)", errors, len(issues))
			}
			return nil
		},
	}
	return cmd
}
