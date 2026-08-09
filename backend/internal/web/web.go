package web

import "embed"

// Dist holds the compiled frontend (index.html, assets, ...).
// Populated by scripts/build.sh copying frontend/dist here.
//
//go:embed all:dist
var Dist embed.FS
