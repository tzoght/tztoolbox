package main

import (
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	"github.com/tzoght/tztoolbox/internal/ui"
)

func newUpdateCmd() *cobra.Command {
	var force bool
	var remote string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Pull the latest tztoolbox from the configured remote (fast-forward only)",
		Long: `Run a fast-forward-only git pull from the named remote (default: origin)
inside the tztoolbox checkout. Refuses to run if the working tree has
uncommitted changes, unless --force is passed.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			root, err := resolveRoot()
			if err != nil {
				return err
			}
			if !force {
				out, err := runGit(root, "status", "--porcelain")
				if err != nil {
					return err
				}
				if strings.TrimSpace(out) != "" {
					return errorf("working tree is dirty; commit/stash changes or rerun with --force\n%s", out)
				}
			}
			out, err := runGit(root, "pull", "--ff-only", remote)
			if err != nil {
				return errorf("git pull failed: %w\n%s", err, out)
			}
			ui.From(cmd.Context()).Raw(out)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "pull even if the working tree is dirty")
	cmd.Flags().StringVar(&remote, "remote", "origin", "remote to pull from")
	return cmd
}

func runGit(dir string, args ...string) (string, error) {
	c := exec.Command("git", args...)
	c.Dir = dir
	out, err := c.CombinedOutput()
	return string(out), err
}
