package main

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/tzoght/tztoolbox/internal/model"
	"github.com/tzoght/tztoolbox/internal/overrides"
	"github.com/tzoght/tztoolbox/internal/render"
	"github.com/tzoght/tztoolbox/internal/ui"
)

func newEnableCmd() *cobra.Command {
	return newToggleCmd("enable", "Enable an artifact for one tool",
		func(root string, t model.Tool, k overrides.Kind, name string) error {
			return overrides.Enable(root, t, k, name)
		})
}

func newDisableCmd() *cobra.Command {
	return newToggleCmd("disable", "Disable an artifact for one tool",
		func(root string, t model.Tool, k overrides.Kind, name string) error {
			return overrides.Disable(root, t, k, name)
		})
}

func newToggleCmd(use, short string, fn func(string, model.Tool, overrides.Kind, string) error) *cobra.Command {
	var toolFlag string
	var kindFlag string
	var skipSync bool
	cmd := &cobra.Command{
		Use:   fmt.Sprintf("%s <name> --tool=<cursor|claude|codex> --kind=<command|rule|skill>", use),
		Short: short,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			t, err := model.ParseTool(toolFlag)
			if err != nil {
				return err
			}
			k, err := overrides.ParseKind(kindFlag)
			if err != nil {
				return err
			}
			root, err := resolveRoot()
			if err != nil {
				return err
			}
			if err := fn(root, t, k, name); err != nil {
				return err
			}
			r := ui.From(cmd.Context())
			r.Note(fmt.Sprintf("%s %s/%s for %s", use, k, name, t))
			if !skipSync {
				if _, err := render.Render(root, render.Options{Tools: []model.Tool{t}}); err != nil {
					return errorf("sync after %s failed: %w", use, err)
				}
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&toolFlag, "tool", "", "target tool: cursor|claude|codex (required)")
	cmd.Flags().StringVar(&kindFlag, "kind", "", "artifact kind: command|rule|skill (required)")
	cmd.Flags().BoolVar(&skipSync, "no-sync", false, "skip running `tzcli sync` afterwards")
	_ = cmd.MarkFlagRequired("tool")
	_ = cmd.MarkFlagRequired("kind")
	return cmd
}
