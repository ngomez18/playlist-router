package static

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var embeddedFiles embed.FS

func GetFrontendFS() (fs.FS, error) {
	return fs.Sub(embeddedFiles, "dist")
}
