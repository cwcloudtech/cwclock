package handlers

import (
	"context"

	"cwclock-api/internal/store"
)

// mailAllowed reports whether an invoice or export-job email may be sent
// for orgID, reserving one unit of its monthly counter if so
// (ai-instruct-83). Organizations owned by a superuser are exempt and never
// counted.
func mailAllowed(ctx context.Context, orgs *store.OrgStore, counters *store.MailCounterStore, orgID string, limit int) (bool, error) {
	isSuperuser, err := orgs.IsOwnedBySuperuser(ctx, orgID)
	if err != nil {
		return false, err
	}
	if isSuperuser {
		return true, nil
	}
	return counters.Reserve(ctx, orgID, limit)
}
