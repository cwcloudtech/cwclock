package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrMailLimitExceeded is returned when an organization's monthly
// invoice/export-job email counter (see MailCounterStore) has already
// reached its limit.
var ErrMailLimitExceeded = errors.New("mail limit exceeded")

type MailCounterStore struct {
	pool *pgxpool.Pool
}

func NewMailCounterStore(pool *pgxpool.Pool) *MailCounterStore {
	return &MailCounterStore{pool: pool}
}

// Reserve atomically counts one more invoice/export-job email against
// orgID's monthly limit (ai-instruct-83). The counter resets to 1 (rather
// than being incremented) when it was last touched before the current
// calendar month. It reports true, having incremented the stored count,
// only when doing so stays within limit; when the limit was already reached
// this month it leaves the counter untouched and reports false, so the
// caller must not send the email.
func (s *MailCounterStore) Reserve(ctx context.Context, orgID string, limit int) (bool, error) {
	var count int
	err := s.pool.QueryRow(ctx, `
		INSERT INTO mail_counter (orga_id, count, updated_at)
		VALUES ($1, 1, now())
		ON CONFLICT (orga_id) DO UPDATE SET
			count = CASE
				WHEN mail_counter.updated_at < date_trunc('month', now()) THEN 1
				ELSE mail_counter.count + 1
			END,
			updated_at = now()
		WHERE mail_counter.updated_at < date_trunc('month', now())
			OR mail_counter.count < $2
		RETURNING count
	`, orgID, limit).Scan(&count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
