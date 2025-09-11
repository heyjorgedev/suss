package dist

import (
	"embed"

	"github.com/benbjohnson/hashfs"
)

//go:embed css/*.css
var fsys embed.FS

var FS = hashfs.NewFS(fsys)
