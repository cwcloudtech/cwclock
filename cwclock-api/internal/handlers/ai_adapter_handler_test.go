package handlers

import (
	"testing"

	"cwclock-api/internal/models"
)

// pickAdapterConnection is the CWAI adapter's subscription picker: the `adapter`
// field of CWCloud's payload matched against the caller's subscriptions — the id
// exactly, the name as a contained, case-insensitive substring.
func TestPickAdapterConnection(t *testing.T) {
	// ListFor order: the caller's own subscriptions first, then the shared ones.
	connections := []models.AIConnection{
		{ID: "11111111-1111-1111-1111-111111111111", Name: "Test"},
		{ID: "22222222-2222-2222-2222-222222222222", Name: "Test (big model)"},
		{ID: "33333333-3333-3333-3333-333333333333", Name: " Claude (daily driver) "},
		{ID: "44444444-4444-4444-4444-444444444444", Name: ""},
	}

	tests := []struct {
		name    string
		adapter string
		want    string
	}{
		{"exact name", "Test", "11111111-1111-1111-1111-111111111111"},
		{"different case", "tEsT", "11111111-1111-1111-1111-111111111111"},
		{"surrounding spaces", "   Test   ", "11111111-1111-1111-1111-111111111111"},
		{"exact name beats a partial one", "test", "11111111-1111-1111-1111-111111111111"},
		{"exact id", "22222222-2222-2222-2222-222222222222", "22222222-2222-2222-2222-222222222222"},
		{"contained fragment", "big", "22222222-2222-2222-2222-222222222222"},
		{"contained fragment across words", "ig mod", "22222222-2222-2222-2222-222222222222"},
		{"contained fragment, different case", "BIG MODEL", "22222222-2222-2222-2222-222222222222"},
		{"contained against a padded name", "claude", "33333333-3333-3333-3333-333333333333"},
		{"contained is the whole padded name", "Claude (daily driver)", "33333333-3333-3333-3333-333333333333"},
		{"no match falls back to the default", "mistral", ""},
		{"an unknown id falls back to the default", "99999999-9999-9999-9999-999999999999", ""},
		{"empty adapter falls back to the default", "", ""},
		{"blank adapter falls back to the default", "   ", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pickAdapterConnection(connections, tt.adapter); got != tt.want {
				t.Errorf("pickAdapterConnection(%q) = %q; want %q", tt.adapter, got, tt.want)
			}
		})
	}

	t.Run("no subscriptions", func(t *testing.T) {
		if got := pickAdapterConnection(nil, "Test"); got != "" {
			t.Errorf("pickAdapterConnection(nil, %q) = %q; want \"\"", "Test", got)
		}
	})

	t.Run("an id wins over another subscription's exact name", func(t *testing.T) {
		// The adapter string is both one subscription's id and another's name:
		// the id is the certain match, so it wins wherever it sits in the list.
		byID := []models.AIConnection{
			{ID: "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", Name: "claude"},
			{ID: "claude", Name: "Claude (daily driver)"},
		}
		if got := pickAdapterConnection(byID, "claude"); got != "claude" {
			t.Errorf("pickAdapterConnection(%q) = %q; want %q", "claude", got, "claude")
		}
	})
}
