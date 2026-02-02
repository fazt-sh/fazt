package help

import "embed"

// Embed the CLI documentation files
// Uses all:cli/* to include subdirectories
//
//go:embed all:cli
var embeddedDocs embed.FS

func init() {
	SetDocs(embeddedDocs)
}
