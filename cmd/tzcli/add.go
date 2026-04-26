package main

import (
	"github.com/spf13/cobra"

	"github.com/tzoght/tztoolbox/internal/authoring"
	"github.com/tzoght/tztoolbox/internal/render"
	"github.com/tzoght/tztoolbox/internal/ui"
)

func newAddCmd() *cobra.Command {
	var skipSync bool
	cmd := &cobra.Command{
		Use:   "add <command|rule|skill> <name>",
		Short: "Scaffold a new artifact under shared/",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			kind, err := authoring.ParseKind(args[0])
			if err != nil {
				return err
			}
			name := args[1]

			root, err := resolveRoot()
			if err != nil {
				return err
			}
			created, err := authoring.Add(root, kind, name)
			if err != nil {
				return err
			}
			r := ui.From(cmd.Context())
			for _, p := range created {
				r.KeyValues(ui.KV{Key: "created", Value: p, Style: ui.StylePath})
			}
			if !skipSync {
				if _, err := render.Render(root, render.Options{}); err != nil {
					return errorf("sync after add failed: %w", err)
				}
				r.Note("synced .cursor/, .claude/, .codex/")
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&skipSync, "no-sync", false, "skip running `tzcli sync` after scaffolding")
	return cmd
}
