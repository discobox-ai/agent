// Package version reports the build's version: the release tag set by
// ldflags, else the VCS revision the binary was built from, else the module
// version Go recorded.
package version

import "runtime/debug"

// Version is set at release time with
// -ldflags "-X github.com/discobox-ai/agent/internal/version.Version=vX.Y.Z".
var Version = ""

// String returns the release version, else the VCS revision with a "-dirty"
// suffix for a modified tree, else the module version Go recorded (as for
// `go run …@main`), else "dev".
func String() string {
	if Version != "" {
		return Version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
	return fromBuildInfo(info)
}

func fromBuildInfo(info *debug.BuildInfo) string {
	var revision, modified string
	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			revision = s.Value
		case "vcs.modified":
			modified = s.Value
		}
	}
	if revision == "" {
		if v := info.Main.Version; v != "" && v != "(devel)" {
			return v
		}
		return "dev"
	}
	if len(revision) > 12 {
		revision = revision[:12]
	}
	if modified == "true" {
		revision += "-dirty"
	}
	return revision
}
