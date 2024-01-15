package web

import "embed"

//go:embed *.html *.js *.png
var Content embed.FS
