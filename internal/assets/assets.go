package assets

import "embed"

//go:embed web/templates/*.html
//go:embed web/static/*
var WebFS embed.FS
