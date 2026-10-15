// Package instance seeds a manager's checkout from the templates and records
// what it wrote, so a later upgrade can merge against it (ADR 26-10-15-157
// §12).
package instance

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	// StateDir holds what discobox-agent writes about the instance.
	StateDir = ".discobox-agent"
	// ManifestPath is the manifest's path in the instance.
	ManifestPath = StateDir + "/manifest.json"
	// BaseDir holds each template as the seeding version wrote it.
	BaseDir = StateDir + "/base"
)

// Manifest records the version that seeded an instance and the templates it
// wrote.
type Manifest struct {
	// Version is the discobox-agent version the base copies came from.
	Version string `json:"version"`
	// Files are the template paths, slash-separated and sorted.
	Files []string `json:"files"`
}

// links are the symlinks every instance has. Init writes them itself and
// skips a template at the same path: embed skips a symlink, but a checkout
// without symlinks (Windows) turns one into a regular file that embed ships.
var links = []struct{ path, target string }{
	{"CLAUDE.md", "AGENTS.md"},
	{".claude/skills", "../.agents/skills"},
}

// Init seeds dir from templates as a new git repository with one commit.
// dir must be missing or an empty directory. On failure Init removes what it
// wrote and every directory it created.
func Init(ctx context.Context, dir string, templates fs.FS, version string) (err error) {
	created, err := prepare(dir)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, cleanup(dir, created))
		}
	}()

	manifest := Manifest{Version: version, Files: []string{}}
	err = fs.WalkDir(templates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || isLink(path) {
			return err
		}
		data, err := fs.ReadFile(templates, path)
		if err != nil {
			return err
		}
		for _, to := range []string{path, BaseDir + "/" + path} {
			if err := writeFile(dir, to, data); err != nil {
				return err
			}
		}
		manifest.Files = append(manifest.Files, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("write templates: %w", err)
	}
	for _, l := range links {
		target := filepath.Join(dir, filepath.FromSlash(l.path))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.Symlink(filepath.FromSlash(l.target), target); err != nil {
			return err
		}
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFile(dir, ManifestPath, append(data, '\n')); err != nil {
		return err
	}

	for _, args := range [][]string{
		{"init", "--quiet", "--initial-branch=main"},
		{"add", "--all"},
		{"commit", "--quiet", "--message", "chore: seed from discobox-agent " + version},
	} {
		if err := git(ctx, dir, args...); err != nil {
			return err
		}
	}
	return nil
}

func isLink(path string) bool {
	for _, l := range links {
		if l.path == path {
			return true
		}
	}
	return false
}

// prepare makes sure dir exists and is empty. It returns the outermost
// directory it had to create, for a failed Init to remove, or "" when dir
// was already there.
func prepare(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		created := dir
		for {
			parent := filepath.Dir(created)
			if _, err := os.Lstat(parent); err == nil || parent == created {
				break
			}
			created = parent
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", err
		}
		return created, nil
	case err != nil:
		return "", err
	case len(entries) > 0:
		return "", fmt.Errorf("%s is not empty; seed a new directory", dir)
	}
	return "", nil
}

// cleanup undoes a failed Init: it removes the directories Init created, or
// else everything Init put in dir.
func cleanup(dir, created string) error {
	if created != "" {
		return os.RemoveAll(created)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	var errs []error
	for _, e := range entries {
		errs = append(errs, os.RemoveAll(filepath.Join(dir, e.Name())))
	}
	return errors.Join(errs...)
}

func writeFile(dir, path string, data []byte) error {
	target := filepath.Join(dir, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, data, 0o644) //nolint:gosec // a checkout's files, as git writes them
}

func git(ctx context.Context, dir string, args ...string) error {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git %s: %w: %s", args[0], err, bytes.TrimSpace(out.Bytes()))
	}
	return nil
}
