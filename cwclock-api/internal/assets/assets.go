// Package assets embeds CWClock's static binary assets so they can be
// shared across packages (report PDFs, transactional emails, ...) without
// each one needing its own copy or a go:embed reaching outside its
// directory tree.
package assets

import _ "embed"

//go:embed cwclock-logo.png
var CWClockLogoPNG []byte
