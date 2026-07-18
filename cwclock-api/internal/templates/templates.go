// Package templates embeds CWClock's shared Go templates (report headers,
// transactional emails) so they live in their own reviewable files instead
// of Go string literals, and so they can be shared across packages without
// a go:embed reaching outside its own directory tree.
package templates

import _ "embed"

//go:embed header.tpl.md
var HeaderMarkdown string

//go:embed email.tpl.html
var EmailHTML string
