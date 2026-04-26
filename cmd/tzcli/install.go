package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tzoght/tztoolbox/internal/installer"
	"github.com/tzoght/tztoolbox/internal/ui"
)

func newInstallCmd() *cobra.Command {
	var only string
	var dryRun bool
	var noPrune bool
	var homeDir string

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Copy the in-repo native trees into ~/.cursor, ~/.claude, ~/.codex",
		Long: `Install the rendered artifacts into your tool home directories.

By default, install only targets tools that are detected on this machine
(i.e. their home directory exists). Use --only cursor|claude|codex to
force installation for a single tool regardless of detection.

Each install pass tracks the set of installed paths in a manifest file
inside each tool home; reinstalling cleans up paths removed upstream
(disable with --no-prune).`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot()
			if err != nil {
				return err
			}
			tools, err := parseToolList(only)
			if err != nil {
				return err
			}

			opts := installer.Options{
				HomeDir: homeDir,
				DryRun:  dryRun,
				Prune:   !noPrune,
			}
			if only != "" && only != "all" {
				opts.Tools = tools
			}

			res, err := installer.Install(root, opts)
			if err != nil {
				return err
			}
			if flagJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(res)
			}
			r := ui.From(cmd.Context())
			prefix := ""
			if dryRun {
				prefix = "[dry-run] "
			}
			rows := make([][]ui.Cell, 0, len(tools))
			for _, t := range tools {
				pr := res.PerTool[t]
				if !pr.Installed {
					rows = append(rows, []ui.Cell{
						ui.StyledCell(prefix+t.String(), ui.StyleTool),
						ui.StyledCell("skipped", ui.StyleMuted),
						ui.StyledCell(pr.Skipped, ui.StyleMuted),
						ui.PlainCell("-"),
						ui.PlainCell("-"),
					})
					continue
				}
				rows = append(rows, []ui.Cell{
					ui.StyledCell(prefix+t.String(), ui.StyleTool),
					ui.StyledCell("ok", ui.StyleSuccess),
					ui.StyledCell(pr.HomePath, ui.StylePath),
					ui.PlainCell(fmt.Sprintf("%d", pr.Copied)),
					ui.PlainCell(fmt.Sprintf("%d", pr.Pruned)),
				})
			}
			r.Table([]string{"TOOL", "RESULT", "HOME", "COPIED", "PRUNED"}, rows)
			return nil
		},
	}
	cmd.Flags().StringVar(&only, "only", "", "install only one tool: cursor|claude|codex|all (default: all detected)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "report the work without writing")
	cmd.Flags().BoolVar(&noPrune, "no-prune", false, "do not remove paths from previous installs that are no longer in the source tree")
	cmd.Flags().StringVar(&homeDir, "home", "", "override the home directory (mostly for tests)")
	return cmd
}
