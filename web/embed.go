// Package web provides embedded static assets and templates.
package web

import "embed"

// Assets contains embedded static files and templates.
//
//go:embed all:static all:templates
var Assets embed.FS
