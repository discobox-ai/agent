package instance

import (
	"context"
	"encoding/json"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/discobox-ai/agent/internal/templates"
)

// TestMain gives git a fixed identity and no user or system config, so Init
// commits the same way on every machine.
func TestMain(m *testing.M) {
	for k, v := range map[string]string{
		"GIT_AUTHOR_NAME":     "test",
		"GIT_AUTHOR_EMAIL":    "test@example.com",
		"GIT_COMMITTER_NAME":  "test",
		"GIT_COMMITTER_EMAIL": "test@example.com",
		"GIT_CONFIG_GLOBAL":   os.DevNull,
		"GIT_CONFIG_NOSYSTEM": "1",
	} {
		if err := os.Setenv(k, v); err != nil {
			panic(err)
		}
	}
	os.Exit(m.Run())
}

var tree = fstest.MapFS{
	"AGENTS.md":                       {Data: []byte("# manager\n")},
	".agents/skills/setup/SKILL.md":   {Data: []byte("setup\n")},
	".agents/worker-skills/README.md": {Data: []byte("catalog\n")},
}

func TestInitSeedsACommittedInstance(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(t.TempDir(), "manager")
	if err := Init(t.Context(), dir, tree, "v1.2.3"); err != nil {
		t.Fatal(err)
	}

	for path, f := range tree {
		for _, at := range []string{path, BaseDir + "/" + path} {
			got, err := os.ReadFile(filepath.Join(dir, at))
			if err != nil || string(got) != string(f.Data) {
				t.Errorf("%s = %q, %v; want %q", at, got, err, f.Data)
			}
		}
	}
	for _, l := range links {
		got, err := os.Readlink(filepath.Join(dir, l.path))
		if err != nil || got != filepath.FromSlash(l.target) {
			t.Errorf("%s -> %q, %v; want -> %q", l.path, got, err, l.target)
		}
	}
	if got, err := os.ReadFile(filepath.Join(dir, "CLAUDE.md")); err != nil || string(got) != "# manager\n" {
		t.Errorf("CLAUDE.md reads %q, %v; want AGENTS.md", got, err)
	}

	data, err := os.ReadFile(filepath.Join(dir, ManifestPath))
	if err != nil {
		t.Fatal(err)
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	want := []string{".agents/skills/setup/SKILL.md", ".agents/worker-skills/README.md", "AGENTS.md"}
	if m.Version != "v1.2.3" || !slices.Equal(m.Files, want) {
		t.Errorf("manifest = %+v; want version v1.2.3, files %v", m, want)
	}

	if out := gitOut(t, dir, "status", "--porcelain"); out != "" {
		t.Errorf("tree not clean after Init:\n%s", out)
	}
	if out := gitOut(t, dir, "log", "--format=%s"); out != "chore: seed from discobox-agent v1.2.3" {
		t.Errorf("history = %q; want one seed commit", out)
	}
	if out := gitOut(t, dir, "branch", "--show-current"); out != "main" {
		t.Errorf("branch = %q; want main", out)
	}
}

func TestInitLinksWhereATemplateHasASymlinkPath(t *testing.T) {
	t.Parallel()
	// A checkout without symlinks turns the template tree's CLAUDE.md into a
	// regular file holding its target, and embed ships it.
	withLink := fstest.MapFS{"CLAUDE.md": {Data: []byte("AGENTS.md")}}
	maps.Copy(withLink, tree)
	dir := t.TempDir()
	if err := Init(t.Context(), dir, withLink, "v1"); err != nil {
		t.Fatal(err)
	}
	if got, err := os.Readlink(filepath.Join(dir, "CLAUDE.md")); err != nil || got != "AGENTS.md" {
		t.Errorf("CLAUDE.md -> %q, %v; want a symlink to AGENTS.md", got, err)
	}
	if _, err := os.Lstat(filepath.Join(dir, BaseDir, "CLAUDE.md")); !os.IsNotExist(err) {
		t.Errorf("CLAUDE.md has a base copy: %v", err)
	}
}

func TestInitSeedsAnExistingEmptyDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := Init(t.Context(), dir, tree, "v1"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "AGENTS.md")); err != nil {
		t.Error(err)
	}
}

func TestInitRefusesANonEmptyDirectory(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	keep := filepath.Join(dir, "notes.txt")
	if err := os.WriteFile(keep, []byte("mine"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := Init(t.Context(), dir, tree, "v1")
	if err == nil || !strings.Contains(err.Error(), "not empty") {
		t.Fatalf("Init = %v; want a not-empty error", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("Init changed a non-empty directory: %v", entries)
	}
}

func TestInitRemovesWhatItWroteOnFailure(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(t.Context())
	cancel() // git cannot start, after the files are written.
	for _, tc := range []struct {
		name     string
		existing bool
		path     string
	}{
		{"missing", false, "manager"},
		{"missing parents", false, "a/b/manager"},
		{"existing empty", true, "manager"},
	} {
		root := t.TempDir()
		dir := filepath.Join(root, filepath.FromSlash(tc.path))
		if tc.existing {
			if err := os.Mkdir(dir, 0o755); err != nil {
				t.Fatal(err)
			}
		}
		if err := Init(ctx, dir, tree, "v1"); err == nil {
			t.Fatalf("%s: Init with a canceled context succeeded", tc.name)
		}
		var left []string
		_ = filepath.WalkDir(root, func(path string, _ os.DirEntry, _ error) error {
			if path != root && (path != dir || !tc.existing) {
				left = append(left, path)
			}
			return nil
		})
		if len(left) > 0 {
			t.Errorf("%s: Init left %v", tc.name, left)
		}
	}
}

func TestInitSeedsTheShippedTemplates(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := Init(t.Context(), dir, templates.Instance(), "dev"); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"CLAUDE.md", ".claude/skills/setup/SKILL.md", ".agents/worker-skills/README.md"} {
		if _, err := os.Stat(filepath.Join(dir, path)); err != nil {
			t.Error(err)
		}
	}
}

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.CommandContext(t.Context(), "git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}
