package main

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tzoght/tztoolbox/internal/inventory"
	"github.com/tzoght/tztoolbox/internal/model"
	"github.com/tzoght/tztoolbox/internal/ui"
)

func newInstalledCmd() *cobra.Command {
	var toolFlag string
	var kindFlag string
	var managedOnly bool
	var homeDir string

	cmd := &cobra.Command{
		Use:   "installed",
		Short: "List artifacts deployed under ~/.cursor, ~/.claude, ~/.codex (side-by-side with the in-repo trees)",
		Long: `Enumerate what is actually installed under each tool's home directory,
side-by-side with the in-repo native trees. Each row tells you whether
the artifact is in the repo, in the home, listed in the install manifest,
and whether it is in sync.

The 'status' field takes one of:
  in_sync       same bytes in repo and home
  only_in_repo  present in <repo>/.<tool> but not installed
  only_in_home  present in ~/.<tool> but absent upstream (likely orphan)
  changed       present in both, contents differ`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot()
			if err != nil {
				return err
			}

			tools, err := parseToolList(toolFlag)
			if err != nil {
				return err
			}

			var kinds []inventory.Kind
			if kindFlag != "" {
				k, err := inventory.ParseKind(kindFlag)
				if err != nil {
					return err
				}
				kinds = []inventory.Kind{k}
			}

			rep, err := inventory.Scan(root, homeDir, tools, inventory.Options{
				ManagedOnly: managedOnly,
				Kinds:       kinds,
			})
			if err != nil {
				return err
			}

			if flagJSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(rep)
			}
			return printInstalled(cmd, rep, tools)
		},
	}
	cmd.Flags().StringVar(&toolFlag, "tool", "all", "limit to one tool: cursor|claude|codex|all")
	cmd.Flags().StringVar(&kindFlag, "kind", "", "limit to one artifact kind: command|rule|skill (default: all)")
	cmd.Flags().BoolVar(&managedOnly, "managed", false, "only show artifacts listed in the home directory's .tztoolbox.manifest")
	cmd.Flags().StringVar(&homeDir, "home", "", "override the home directory (mostly for tests)")
	return cmd
}

func printInstalled(cmd *cobra.Command, rep inventory.Report, tools []model.Tool) error {
	r := ui.From(cmd.Context())
	for i, t := range tools {
		entry := rep.Tools[t]
		if i > 0 {
			r.Note("")
		}
		homeStyle := ui.StyleSuccess
		homeMark := "exists"
		if !entry.HomeExists {
			homeMark = "missing"
			homeStyle = ui.StyleWarn
		}
		r.Heading(fmt.Sprintf("Tool: %s", t))
		r.KeyValues(
			ui.KV{Key: "home", Value: entry.HomeDir + " [" + homeMark + "]", Style: homeStyle},
			ui.KV{Key: "repo", Value: entry.RepoDir, Style: ui.StylePath},
		)
		if len(entry.Items) == 0 {
			r.Note("(no entries)")
			continue
		}
		rows := make([][]ui.Cell, 0, len(entry.Items))
		for _, it := range entry.Items {
			rows = append(rows, []ui.Cell{
				ui.PlainCell(string(it.Kind)),
				ui.StyledCell(it.Name, ui.StyleBold),
				ynCell(it.Repo),
				ynCell(it.Home),
				ynCell(it.Managed),
				statusCell(it.Status),
			})
		}
		r.Table(
			[]string{"KIND", "NAME", "REPO", "HOME", "MANAGED", "STATUS"},
			rows,
		)
	}
	return nil
}

// ynCell renders a yes/no value with muted style for "no".
func ynCell(b bool) ui.Cell {
	if b {
		return ui.StyledCell("yes", ui.StyleSuccess)
	}
	return ui.StyledCell("-", ui.StyleMuted)
}

// statusCell colors the status field according to its semantic meaning.
func statusCell(s inventory.Status) ui.Cell {
	switch s {
	case inventory.StatusInSync:
		return ui.StyledCell(string(s), ui.StyleSuccess)
	case inventory.StatusChanged:
		return ui.StyledCell(string(s), ui.StyleWarn)
	case inventory.StatusOnlyInRepo, inventory.StatusOnlyInHome:
		return ui.StyledCell(string(s), ui.StyleMuted)
	}
	return ui.PlainCell(string(s))
}
