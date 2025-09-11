package dist

import (
	"embed"

	"github.com/benbjohnson/hashfs"
)

//go:embed css/*.css
//go:embed favicon/*
var fsys embed.FS

var FS = hashfs.NewFS(fsys)
