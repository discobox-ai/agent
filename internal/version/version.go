// Package version reports the build's version: the release tag set by
// ldflags, or the VCS revision the binary was built from.
package version

import "runtime/debug"

// Version is set at release time with
// -ldflags "-X github.com/discobox-ai/agent/internal/version.Version=vX.Y.Z".
var Version = ""

// String returns the release version, else the VCS revision with a "-dirty"
// suffix for a modified tree, else "dev".
func String() string {
	if Version != "" {
		return Version
	}
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "dev"
	}
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
