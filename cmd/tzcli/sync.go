package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tzoght/tztoolbox/internal/render"
	"github.com/tzoght/tztoolbox/internal/ui"
)

func newSyncCmd() *cobra.Command {
	var checkOnly bool
	var dryRun bool
	var only string

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Render shared/ + overrides/ into the in-repo .cursor, .claude, .codex trees",
		Long: `Render the canonical artifacts under shared/ (with per-tool overrides
under overrides/<tool>/) into the in-repo native trees: .cursor/,
.claude/, .codex/.

Idempotent: running sync repeatedly on unchanged inputs produces the
same bytes. Use --check to make sync non-destructive: it exits with a
non-zero status if rendering would change anything.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot()
			if err != nil {
				return err
			}
			tools, err := parseToolList(only)
			if err != nil {
				return err
			}

			r := ui.From(cmd.Context())

			if checkOnly {
				report, err := render.CheckDrift(root, tools)
				if err != nil {
					return err
				}
				if flagJSON {
					return json.NewEncoder(cmd.OutOrStdout()).Encode(report)
				}
				if !report.HasDrift() {
					r.Note("no drift: in-repo trees match shared/ + overrides/")
					return nil
				}
				r.Warn(fmt.Sprintf("drift detected (%d entries)", len(report.Entries)))
				rows := make([][]ui.Cell, 0, len(report.Entries))
				for _, e := range report.Entries {
					rows = append(rows, []ui.Cell{
						ui.StyledCell(string(e.Tool), ui.StyleTool),
						ui.PlainCell(string(e.Kind)),
						ui.StyledCell(e.Path, ui.StylePath),
					})
				}
				r.Table([]string{"TOOL", "KIND", "PATH"}, rows)
				return errorf("run `tzcli sync` to update the in-repo trees")
			}

			res, err := render.Render(root, render.Options{Tools: tools, DryRun: dryRun})
			if err != nil {
				return err
			}
			if flagJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(res)
			}
			prefix := ""
			if dryRun {
				prefix = "[dry-run] "
			}
			rows := make([][]ui.Cell, 0, len(tools))
			for _, t := range tools {
				c := res.PerTool[t]
				rows = append(rows, []ui.Cell{
					ui.StyledCell(prefix+t.String(), ui.StyleTool),
					ui.PlainCell(fmt.Sprintf("%d", c.Commands)),
					ui.PlainCell(fmt.Sprintf("%d", c.Rules)),
					ui.PlainCell(fmt.Sprintf("%d", c.Skills)),
				})
			}
			r.Table([]string{"TOOL", "COMMANDS", "RULES", "SKILLS"}, rows)
			return nil
		},
	}
	cmd.Flags().BoolVar(&checkOnly, "check", false, "report drift without writing; non-zero exit if differences exist")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report what would be rendered without writing")
	cmd.Flags().StringVar(&only, "only", "all", "render only one tool: cursor|claude|codex|all")
	return cmd
}
