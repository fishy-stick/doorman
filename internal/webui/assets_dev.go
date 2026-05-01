//go:build !embedweb

package webui

import "io/fs"

func assetFS() (fs.FS, bool) {
	return nil, false
}
