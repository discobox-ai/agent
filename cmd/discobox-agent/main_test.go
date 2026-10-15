package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRejectsBadArguments(t *testing.T) {
	t.Parallel()
	for _, args := range [][]string{nil, {"init"}, {"init", "a", "b"}, {"upgrade"}} {
		var stdout, stderr bytes.Buffer
		if code := run(t.Context(), args, &stdout, &stderr); code != 2 || !strings.Contains(stderr.String(), "usage:") {
			t.Errorf("run(%q) = %d, stderr %q; want 2 and usage", args, code, stderr.String())
		}
	}
}

func TestRunInitReportsAFailure(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if code := run(t.Context(), []string{"init", dir}, &stdout, &stderr); code != 1 || !strings.Contains(stderr.String(), "not empty") {
		t.Errorf("run(init non-empty) = %d, stderr %q; want 1 and the reason", code, stderr.String())
	}
}
