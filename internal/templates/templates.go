// Package templates embeds the files discobox-agent seeds a manager with.
package templates

import (
	"embed"
	"io/fs"
)

// The "all:" prefix keeps the dot-directories (.agents) that embed would
// otherwise skip.
//
//go:embed all:instance
var embedded embed.FS

// Instance returns the template tree, rooted as a manager's checkout is.
func Instance() fs.FS {
	sub, err := fs.Sub(embedded, "instance")
	if err != nil {
		panic(err) // "instance" is a valid path, so this cannot fail.
	}
	return sub
}
