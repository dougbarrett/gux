package defaults

import "embed"

//go:embed index.html manifest.json service-worker.js
var Files embed.FS
