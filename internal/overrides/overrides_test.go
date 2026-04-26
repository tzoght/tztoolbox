package overrides

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tzoght/tztoolbox/internal/model"
)

func TestEnableDisableRoundTrip(t *testing.T) {
	root := t.TempDir()

	if err := Disable(root, model.Claude, KindCommand, "foo"); err != nil {
		t.Fatalf("Disable: %v", err)
	}
	d, err := Load(root, model.Claude)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(d.Commands) != 1 || d.Commands[0] != "foo" {
		t.Fatalf("expected [foo], got %+v", d.Commands)
	}

	// Disable is idempotent: re-disabling does not duplicate entries.
	if err := Disable(root, model.Claude, KindCommand, "foo"); err != nil {
		t.Fatalf("Disable (2): %v", err)
	}
	d, _ = Load(root, model.Claude)
	if len(d.Commands) != 1 {
		t.Fatalf("idempotency violated; got %v", d.Commands)
	}

	if err := Enable(root, model.Claude, KindCommand, "foo"); err != nil {
		t.Fatalf("Enable: %v", err)
	}
	d, _ = Load(root, model.Claude)
	if len(d.Commands) != 0 {
		t.Fatalf("expected empty after enable, got %v", d.Commands)
	}

	// File should be removed entirely once empty.
	p := filepath.Join(root, "overrides", "claude", "disabled.yaml")
	if _, err := os.Stat(p); err == nil {
		t.Errorf("expected disabled.yaml to be removed when empty")
	}
}
