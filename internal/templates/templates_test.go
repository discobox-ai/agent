package templates

import (
	"errors"
	"io/fs"
	"testing"
)

func TestInstanceHoldsTheManagerUnconfigured(t *testing.T) {
	t.Parallel()
	tree := Instance()
	for _, path := range []string{
		"AGENTS.md",
		".agents/skills/setup/SKILL.md",
		".agents/skills/discobox-cli/SKILL.md",
		".agents/worker-skills/README.md",
		".agents/worker-skills/project/SKILL.md",
	} {
		if _, err := fs.Stat(tree, path); err != nil {
			t.Errorf("template %s: %v", path, err)
		}
	}
	// The manager knows it is unconfigured because agent.yaml is missing.
	if _, err := fs.Stat(tree, "agent.yaml"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("agent.yaml must not ship in the templates, got %v", err)
	}
}
