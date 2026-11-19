// Package assistant embeds the cwclock AI assistant's skill — the model's
// instructions, as markdown files rather than Go string literals, because they are
// meant to be tuned: how the assistant answers and which endpoints it reads are
// judgements that get revised, and revising them should not mean reading Go.
//
// The system prompt a chat is answered under is Skill + "\n\n" + accountContext +
// "\n\n" + languageRule(locale); Skill is the two files below, joined.
package assistant

import (
	_ "embed"
	"strings"
)

//go:embed assistant.skill.md
var AssistantSkill string

//go:embed api.skill.md
var APISkill string

// Skill is the whole assistant skill: the persona and scope, then how to read
// cwclock's API directly (as a tool), in that order.
var Skill = strings.Join([]string{AssistantSkill, APISkill}, "\n\n")
