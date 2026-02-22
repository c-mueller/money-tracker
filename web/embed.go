package web

import "embed"

//go:embed templates static
var Content embed.FS
