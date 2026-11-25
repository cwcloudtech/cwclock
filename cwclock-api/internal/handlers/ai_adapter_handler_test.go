package handlers

import (
	"strings"
	"testing"

	"cwclock-api/internal/assistant"
	"cwclock-api/internal/models"
)

// The subscriptions the adapter picker chooses from, in ListFor order (the
// caller's own first, then the ones shared with them).
const (
	idBigModel      = "11111111-1111-1111-1111-111111111111"
	idDriver        = "22222222-2222-2222-2222-222222222222"
	idDailyDriver   = "33333333-3333-3333-3333-333333333333"
	idUnnamed       = "44444444-4444-4444-4444-444444444444"
	idNotSubscribed = "99999999-9999-9999-9999-999999999999"
)

func adapterConnections() []models.AIConnection {
	return []models.AIConnection{
		{ID: idBigModel, Name: "Big model"},
		{ID: idDriver, Name: "Driver"},
		{ID: idDailyDriver, Name: " Daily driver "},
		{ID: idUnnamed, Name: ""},
	}
}

// pickAdapterConnection is the CWAI adapter's subscription picker. The
// containment runs adapter-side: the `adapter` field CWCloud sends is a label
// that must carry the subscription's name or id (case-insensitive), not a
// fragment a subscription name happens to contain.
func TestPickAdapterConnection(t *testing.T) {
	tests := []struct {
		name    string
		adapter string
		want    string
	}{
		// The adapter IS the name, or the id.
		{"exact name", "Big model", idBigModel},
		{"different case", "big model", idBigModel},
		{"surrounding spaces", "  BIG MODEL  ", idBigModel},
		{"exact id", idDriver, idDriver},
		{"a padded name, exactly", "Daily driver", idDailyDriver},

		// The adapter CARRIES the name — a label written in CWCloud's catalogue.
		{"a label carrying the name", "cwclock big model", idBigModel},
		{"a label carrying the name in brackets", "Sourcing (BIG MODEL)", idBigModel},
		{"a label carrying the name mid-sentence", "ask the big model please", idBigModel},
		{"a label carrying the id", "cwclock-" + idDriver, idDriver},
		{"a label carrying the id, upper case", "CWCLOCK-" + strings.ToUpper(idDriver), idDriver},

		// The other direction must NOT match: a fragment of a name is not a
		// label for it.
		{"a fragment of the name does not match", "big", ""},
		{"a shorter fragment does not match", "mod", ""},
		{"a truncated name does not match", "Daily dri", ""},

		// The longest name the label carries is the one it means.
		{"the longest name carried wins", "ask the daily driver please", idDailyDriver},
		{"only the shorter name is carried", "ask the driver please", idDriver},

		// Nothing matched: the default subscription answers.
		{"no match falls back to the default", "mistral", ""},
		{"an unknown id falls back to the default", idNotSubscribed, ""},
		{"empty adapter falls back to the default", "", ""},
		{"blank adapter falls back to the default", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pickAdapterConnection(adapterConnections(), tt.adapter); got != tt.want {
				t.Errorf("pickAdapterConnection(%q) = %q; want %q", tt.adapter, got, tt.want)
			}
		})
	}

	t.Run("no subscriptions", func(t *testing.T) {
		if got := pickAdapterConnection(nil, "Big model"); got != "" {
			t.Errorf("pickAdapterConnection(nil, %q) = %q; want \"\"", "Big model", got)
		}
	})

	t.Run("a blank name is carried by nothing", func(t *testing.T) {
		// "" is contained in every string, so an unnamed subscription would
		// otherwise answer every adapter CWCloud ever registers.
		for _, adapter := range []string{"whatever", "cwclock", idNotSubscribed} {
			if got := pickAdapterConnection(adapterConnections(), adapter); got == idUnnamed {
				t.Errorf("pickAdapterConnection(%q) picked the unnamed subscription", adapter)
			}
		}
	})

	// The tiers, from the most certain to the least.
	t.Run("the exact id wins over another subscription's exact name", func(t *testing.T) {
		connections := []models.AIConnection{
			{ID: idBigModel, Name: "claude"},
			{ID: "claude", Name: "Daily driver"},
		}
		if got := pickAdapterConnection(connections, "claude"); got != "claude" {
			t.Errorf("pickAdapterConnection(%q) = %q; want %q", "claude", got, "claude")
		}
	})

	t.Run("an exact name wins over a carried id", func(t *testing.T) {
		connections := []models.AIConnection{
			{ID: "big", Name: "Something else"},
			{ID: idBigModel, Name: "cwclock big model"},
		}
		if got := pickAdapterConnection(connections, "cwclock big model"); got != idBigModel {
			t.Errorf("pickAdapterConnection(%q) = %q; want %q", "cwclock big model", got, idBigModel)
		}
	})

	t.Run("a carried id wins over a carried name", func(t *testing.T) {
		connections := []models.AIConnection{
			{ID: idBigModel, Name: "Big model"},
			{ID: idDriver, Name: "Driver"},
		}
		adapter := "big model via " + idDriver
		if got := pickAdapterConnection(connections, adapter); got != idDriver {
			t.Errorf("pickAdapterConnection(%q) = %q; want %q", adapter, got, idDriver)
		}
	})

	t.Run("a name holding a wildcard is matched as text", func(t *testing.T) {
		// Nothing here is a SQL pattern: % and _ are characters.
		connections := []models.AIConnection{
			{ID: idBigModel, Name: "100%_model"},
			{ID: idDriver, Name: "Driver"},
		}
		if got := pickAdapterConnection(connections, "cwclock 100%_model"); got != idBigModel {
			t.Errorf("a literal %% / _ in a name must match as itself, got %q", got)
		}
		if got := pickAdapterConnection(connections, "cwclock 100XYmodel"); got != "" {
			t.Errorf("%% / _ must not act as wildcards, got %q", got)
		}
	})
}

// adapterSystem is the system prompt an adapter question is answered under: the
// chat's own prompt (skill + account context + language rule), then the rule about
// answering through an external integration, and only then whatever `system` turns
// CWCloud sent — appended, and subordinate to cwclock's own.
func TestAdapterSystem(t *testing.T) {
	const context = "# Your CWClock account\n\n- Today's date is 2026-11-25."
	const caller = "You are CWCloud's helpdesk bot. Be terse."

	got := adapterSystem(context, "fr", caller)

	if !strings.HasPrefix(got, assistant.Skill) {
		t.Error("the cwclock skill must come first, so CWCloud's own prompt can never replace it")
	}
	if !strings.Contains(got, context) {
		t.Error("the account context is missing — the model would not know who is asking")
	}
	if !strings.Contains(got, chatLanguageRule("fr")) {
		t.Error("the language rule is missing")
	}
	if !strings.Contains(got, adapterIntegrationRule) {
		t.Error("the external-integration rule is missing")
	}
	if !strings.Contains(got, "downloadUrl") || !strings.Contains(got, "[filename](downloadUrl)") {
		t.Error("the rule must tell the model to hand a generated file over as the exact downloadUrl")
	}
	if !strings.Contains(got, caller) {
		t.Error("CWCloud's own system turns are missing")
	}
	if !strings.Contains(got, adapterCallerSystemHeading+"\n\n"+caller) {
		t.Error("CWCloud's system turns must sit under the heading that subordinates them")
	}
	if strings.Index(got, caller) < strings.Index(got, adapterIntegrationRule) {
		t.Error("CWCloud's system turns must come after cwclock's own rules, not before")
	}

	t.Run("no caller instructions", func(t *testing.T) {
		got := adapterSystem(context, "en", "   ")
		if strings.Contains(got, adapterCallerSystemHeading) {
			t.Error("the caller-instructions heading must be left out when CWCloud sent no system turn")
		}
		if !strings.HasSuffix(got, adapterIntegrationRule) {
			t.Error("the integration rule should then be the last block")
		}
	})

	t.Run("no account context", func(t *testing.T) {
		got := adapterSystem("", "", "")
		if !strings.HasPrefix(got, assistant.Skill+"\n\n"+chatLanguageRule("")) {
			t.Error("an empty context must be skipped, not joined as a blank block")
		}
		// One blank line between the blocks adapterSystem joins (the embedded
		// skill has its own blank lines, so only the joints are checked).
		if !strings.Contains(got, chatLanguageRule("")+"\n\n"+adapterIntegrationRule) {
			t.Error("blocks are joined by exactly one blank line")
		}
	})
}
