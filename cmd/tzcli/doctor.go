package main

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	"github.com/tzoght/tztoolbox/internal/doctor"
	"github.com/tzoght/tztoolbox/internal/model"
	"github.com/tzoght/tztoolbox/internal/ui"
)

func newDoctorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Print environment readiness, detected tools, and drift status",
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot()
			if err != nil {
				return err
			}
			rep, err := doctor.Run(root)
			if err != nil {
				return err
			}
			if flagJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(rep)
			}
			r := ui.From(cmd.Context())

			// Repo section
			r.Heading("repo")
			dirtyStyle := ui.StyleSuccess
			if rep.Repo.IsDirty {
				dirtyStyle = ui.StyleWarn
			}
			r.KeyValues(
				ui.KV{Key: "root", Value: rep.Repo.Root, Style: ui.StylePath},
				ui.KV{Key: "branch", Value: strOrDash(rep.Repo.Branch)},
				ui.KV{Key: "origin", Value: strOrDash(rep.Repo.OriginURL), Style: ui.StyleMuted},
				ui.KV{Key: "dirty", Value: fmt.Sprintf("%t", rep.Repo.IsDirty), Style: dirtyStyle},
			)
			r.Note("")

			// Tools section
			r.Heading("tools")
			toolRows := make([][]ui.Cell, 0, len(model.AllTools))
			for _, t := range model.AllTools {
				ts := rep.Tools[t]
				detectedCell := ui.StyledCell("not detected", ui.StyleMuted)
				if ts.Detected {
					detectedCell = ui.StyledCell("detected", ui.StyleSuccess)
				}
				existsCell := ui.StyledCell("missing", ui.StyleMuted)
				if ts.HomeExists {
					existsCell = ui.StyledCell("exists", ui.StyleSuccess)
				}
				toolRows = append(toolRows, []ui.Cell{
					ui.StyledCell(t.String(), ui.StyleTool),
					detectedCell,
					ui.StyledCell(ts.HomePath, ui.StylePath),
					existsCell,
					ui.PlainCell(strOrDash(ts.Version)),
				})
			}
			r.Table([]string{"TOOL", "DETECTED", "HOME", "EXISTS", "VERSION"}, toolRows)
			r.Note("")

			// Artifact counts
			r.Heading("artifacts (shared/)")
			artRows := make([][]ui.Cell, 0, len(model.AllTools))
			for _, t := range model.AllTools {
				c := rep.Counts[t.String()]
				artRows = append(artRows, []ui.Cell{
					ui.StyledCell(t.String(), ui.StyleTool),
					ui.PlainCell(fmt.Sprintf("%d", c["commands"])),
					ui.PlainCell(fmt.Sprintf("%d", c["rules"])),
					ui.PlainCell(fmt.Sprintf("%d", c["skills"])),
				})
			}
			r.Table([]string{"TOOL", "COMMANDS", "RULES", "SKILLS"}, artRows)
			r.Note("")

			// Helpers
			r.Heading("helpers")
			names := make([]string, 0, len(rep.Helpers))
			for n := range rep.Helpers {
				names = append(names, n)
			}
			sort.Strings(names)
			helperRows := make([][]ui.Cell, 0, len(names))
			for _, n := range names {
				h := rep.Helpers[n]
				if h.Found {
					helperRows = append(helperRows, []ui.Cell{
						ui.StyledCell(n, ui.StyleBold),
						ui.StyledCell("ok", ui.StyleSuccess),
						ui.PlainCell(h.Version),
					})
				} else {
					helperRows = append(helperRows, []ui.Cell{
						ui.StyledCell(n, ui.StyleBold),
						ui.StyledCell("missing", ui.StyleMuted),
						ui.StyledCell("not on PATH", ui.StyleMuted),
					})
				}
			}
			r.Table([]string{"HELPER", "STATUS", "VERSION"}, helperRows)
			r.Note("")

			// Drift
			if rep.Drift.HasDrift() {
				r.Warn(fmt.Sprintf("drift: %d entries (run `tzcli sync` to fix)", len(rep.Drift.Entries)))
				dRows := make([][]ui.Cell, 0, len(rep.Drift.Entries))
				for _, d := range rep.Drift.Entries {
					dRows = append(dRows, []ui.Cell{
						ui.StyledCell(string(d.Tool), ui.StyleTool),
						ui.PlainCell(string(d.Kind)),
						ui.StyledCell(d.Path, ui.StylePath),
					})
				}
				r.Table([]string{"TOOL", "KIND", "PATH"}, dRows)
			} else {
				r.KeyValues(ui.KV{Key: "drift", Value: "none", Style: ui.StyleSuccess})
			}
			return nil
		},
	}
	return cmd
}

func strOrDash(s string) string {
	if s == "" {
		return "-"
	}
	return s
}
