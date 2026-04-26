package main

import (
	"fmt"
	"os"

	"github.com/tzoght/tztoolbox/internal/model"
	"github.com/tzoght/tztoolbox/internal/repo"
)

// resolveRoot returns the tztoolbox repo root, honoring --repo if set and
// otherwise discovering upward from the current working directory.
func resolveRoot() (string, error) {
	if flagRepoRoot != "" {
		return flagRepoRoot, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return repo.Discover(cwd)
}

// parseToolList expands "all" or empty to all tools, otherwise validates each.
func parseToolList(s string) ([]model.Tool, error) {
	if s == "" || s == "all" {
		return model.AllTools, nil
	}
	t, err := model.ParseTool(s)
	if err != nil {
		return nil, err
	}
	return []model.Tool{t}, nil
}

// errorf returns an error formatted like fmt.Errorf so cobra prints it
// without the noisy "Usage:" footer.
func errorf(format string, a ...any) error {
	return fmt.Errorf(format, a...)
}
