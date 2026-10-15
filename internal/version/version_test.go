package version

import (
	"runtime/debug"
	"testing"
)

func TestStringReturnsTheReleaseVersionWhenSet(t *testing.T) {
	old := Version
	t.Cleanup(func() { Version = old })
	Version = "v1.2.3"
	if got := String(); got != "v1.2.3" {
		t.Fatalf("String() = %q, want v1.2.3", got)
	}
}

func TestFromBuildInfo(t *testing.T) {
	t.Parallel()
	vcs := []debug.BuildSetting{{Key: "vcs.revision", Value: "0123456789abcdef"}, {Key: "vcs.modified", Value: "true"}}
	for _, tc := range []struct {
		name     string
		main     string
		settings []debug.BuildSetting
		want     string
	}{
		{"checkout build", "v0.0.0-20261015000000-0123456789ab+dirty", vcs, "0123456789ab-dirty"},
		{"go run from the module cache", "v0.0.0-20261015000000-0123456789ab", nil, "v0.0.0-20261015000000-0123456789ab"},
		{"nothing recorded", "(devel)", nil, "dev"},
	} {
		info := &debug.BuildInfo{Main: debug.Module{Version: tc.main}, Settings: tc.settings}
		if got := fromBuildInfo(info); got != tc.want {
			t.Errorf("%s: fromBuildInfo = %q, want %q", tc.name, got, tc.want)
		}
	}
}
