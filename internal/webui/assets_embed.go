//go:build embedweb

package webui

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed dist
var embeddedDist embed.FS

func assetFS() (fs.FS, bool) {
	sub, err := fs.Sub(embeddedDist, "dist")
	if err != nil {
		panic(fmt.Errorf("sub embedded web dist: %w", err))
	}

	return sub, true
}
