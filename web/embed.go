// Package web embeds the built React/Vite UI so the whole application ships as
// a single binary. The dist/ directory is produced by `npm run build`.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var dist embed.FS

// DistFS returns the embedded UI filesystem rooted at the dist/ directory.
func DistFS() fs.FS {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return sub
}
