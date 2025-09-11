package dist

import (
	"embed"

	"github.com/benbjohnson/hashfs"
)

//go:embed css/*.css
//go:embed favicon/*
//go:embed img/*
var fsys embed.FS

var FS = hashfs.NewFS(fsys)
