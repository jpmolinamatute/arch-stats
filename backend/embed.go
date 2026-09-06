package backend

import (
	"embed"
	"fmt"
	"io/fs"
)

// Frontend holds the built Vue 3 SPA assets.
// In development, the embed may be a placeholder; use filesystem serving instead.
//
//go:embed all:frontend
var Frontend embed.FS

// FS returns an fs.FS sub-tree rooted at the embedded "frontend" directory.
func FS() (fs.FS, error) {
	sub, err := fs.Sub(Frontend, "frontend")
	if err != nil {
		return nil, fmt.Errorf("opening embedded frontend subtree: %w", err)
	}
	return sub, nil
}
