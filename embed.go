// Package karakocdn embeds the static assets served by the CDN.
package karakocdn

import "embed"

// Assets holds the assets/ tree (any file type; dotfiles are included so
// that reserved-but-empty directories keep the build working).
//
//go:embed all:assets
var Assets embed.FS
