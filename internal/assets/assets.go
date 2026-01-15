package assets

import "embed"

//go:embed system/**/*
var SystemFS embed.FS

//go:embed stdlib/*.js
var StdlibFS embed.FS

