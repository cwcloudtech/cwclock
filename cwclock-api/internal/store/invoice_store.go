package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"cwclock-api/internal/models"
	"cwclock-api/internal/utils"
)

type InvoiceStore struct {
	pool *pgxpool.Pool
}

func NewInvoiceStore(pool *pgxpool.Pool) *InvoiceStore {
	return &InvoiceStore{pool: pool}
}

type invoiceData struct {
	Number   string  `json:"number"`
	Status   string  `json:"status"`
	TotalHT  float64 `json:"totalHT"`
	TotalVAT float64 `json:"totalVAT"`
	TotalTTC float64 `json:"totalTTC"`
}

// InvoiceFields holds an invoice's editable, non-relational fields, built
// fresh for each invoice-number attempt in Create (see BuildFields there)
// since the rendered PDF embeds the number it's ultimately saved under.
type InvoiceFields struct {
	Status            string
	TotalHT           float64
	TotalVAT          float64
	TotalTTC          float64
	PDF               []byte
	SelectedBeginDate string
	SelectedEndDate   string
}

const dateLayout = "2006-01-02"

func scanInvoice(row pgx.Row) (models.Invoice, error) {
	var inv models.Invoice
	var raw []byte
	var begin, end time.Time
	if err := row.Scan(&inv.ID, &inv.OrganizationID, &inv.ClientID, &raw, &begin, &end, &inv.CreatedAt, &inv.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Invoice{}, ErrNotFound
		}
		return models.Invoice{}, err
	}
	var d invoiceData
	if err := json.Unmarshal(raw, &d); err != nil {
		return models.Invoice{}, err
	}
	inv.Number = d.Number
	inv.Status = models.InvoiceStatus(d.Status)
	inv.TotalHT = d.TotalHT
	inv.TotalVAT = d.TotalVAT
	inv.TotalTTC = d.TotalTTC
	inv.SelectedBeginDate = begin.Format(dateLayout)
	inv.SelectedEndDate = end.Format(dateLayout)
	return inv, nil
}

var nonAlnum = regexp.MustCompile(`[^A-Z0-9]`)

// invoiceNumberPrefix builds the "{CLIENT_NAME_CAPITALIZED}{YYYYMMDD}" an
// invoice number starts with: the client's name, uppercased and stripped of
// everything but letters/digits, followed by the generation date.
func invoiceNumberPrefix(clientName string, generatedAt time.Time) string {
	return nonAlnum.ReplaceAllString(strings.ToUpper(clientName), utils.EMPTY) + generatedAt.Format("20060102")
}

const maxInvoiceNumberAttempts = 1000

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Create allocates the next free invoice number for today under this
// client (prefix+1, prefix+2, ... on a collision) and inserts the invoice,
// retrying on the unique (organization_id, number) constraint rather than a
// check-then-insert, which would race under concurrent invoice generation.
// buildFields is called once per attempt (usually just once) since the
// rendered PDF embeds the invoice number it's ultimately saved under.
func (s *InvoiceStore) Create(ctx context.Context, orgID, clientID, clientName string, buildFields func(number string) (InvoiceFields, error)) (models.Invoice, error) {
	prefix := invoiceNumberPrefix(clientName, time.Now())
	for attempt := 1; attempt <= maxInvoiceNumberAttempts; attempt++ {
		number := fmt.Sprintf("%s%d", prefix, attempt)
		f, err := buildFields(number)
		if err != nil {
			return models.Invoice{}, err
		}

		data, err := json.Marshal(invoiceData{
			Number: number, Status: f.Status,
			TotalHT: f.TotalHT, TotalVAT: f.TotalVAT, TotalTTC: f.TotalTTC,
		})
		if err != nil {
			return models.Invoice{}, err
		}

		row := s.pool.QueryRow(ctx, `
			INSERT INTO invoices (organization_id, client_id, data, pdf, selected_begin_date, selected_end_date)
			VALUES ($1, $2, $3, $4, $5, $6)
			RETURNING id, organization_id, client_id, data, selected_begin_date, selected_end_date, created_at, updated_at
		`, orgID, clientID, data, f.PDF, f.SelectedBeginDate, f.SelectedEndDate)
		inv, err := scanInvoice(row)
		if err == nil {
			return inv, nil
		}
		if !isUniqueViolation(err) {
			return models.Invoice{}, err
		}
	}
	return models.Invoice{}, fmt.Errorf("could not allocate an invoice number for %q", prefix)
}

// PeekNextNumber returns the invoice number that would be allocated right
// now for this client, for display on a preview PDF - best-effort, not
// reserved, since only Create's retry-on-conflict allocation is
// authoritative.
func (s *InvoiceStore) PeekNextNumber(ctx context.Context, orgID, clientName string) (string, error) {
	prefix := invoiceNumberPrefix(clientName, time.Now())
	var count int
	err := s.pool.QueryRow(ctx, `
		SELECT count(*) FROM invoices WHERE organization_id = $1 AND data->>'number' LIKE $2 || '%'
	`, orgID, prefix).Scan(&count)
	if err != nil {
		return utils.EMPTY, err
	}
	return fmt.Sprintf("%s%d", prefix, count+1), nil
}

// List returns an organization's invoices whose selected period falls
// within [start, end] (both "YYYY-MM-DD"), most recent first, optionally
// narrowed to one client.
func (s *InvoiceStore) List(ctx context.Context, orgID, clientID, start, end string) ([]models.Invoice, error) {
	query := `
		SELECT id, organization_id, client_id, data, selected_begin_date, selected_end_date, created_at, updated_at
		FROM invoices
		WHERE organization_id = $1
		  AND selected_begin_date >= $2::date
		  AND selected_end_date <= $3::date
	`
	args := []any{orgID, start, end}
	if utils.IsNotBlank(clientID) {
		args = append(args, clientID)
		query += fmt.Sprintf(" AND client_id = $%d", len(args))
	}
	query += " ORDER BY created_at DESC"

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invoices := []models.Invoice{}
	for rows.Next() {
		inv, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, inv)
	}
	return invoices, rows.Err()
}

func (s *InvoiceStore) FindByID(ctx context.Context, id string) (models.Invoice, error) {
	row := s.pool.QueryRow(ctx, `
		SELECT id, organization_id, client_id, data, selected_begin_date, selected_end_date, created_at, updated_at
		FROM invoices WHERE id = $1
	`, id)
	return scanInvoice(row)
}

// GetPDF returns an invoice's stored PDF bytes and its human-readable
// number (used as the download filename).
func (s *InvoiceStore) GetPDF(ctx context.Context, id string) (pdf []byte, number string, err error) {
	row := s.pool.QueryRow(ctx, `SELECT pdf, data->>'number' FROM invoices WHERE id = $1`, id)
	if err := row.Scan(&pdf, &number); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, utils.EMPTY, ErrNotFound
		}
		return nil, utils.EMPTY, err
	}
	return pdf, number, nil
}

func (s *InvoiceStore) UpdateStatus(ctx context.Context, id, status string) (models.Invoice, error) {
	row := s.pool.QueryRow(ctx, `
		UPDATE invoices SET data = jsonb_set(data, '{status}', to_jsonb($2::text), true), updated_at = now()
		WHERE id = $1
		RETURNING id, organization_id, client_id, data, selected_begin_date, selected_end_date, created_at, updated_at
	`, id, status)
	return scanInvoice(row)
}

func (s *InvoiceStore) Delete(ctx context.Context, id string) error {
	tag, err := s.pool.Exec(ctx, `DELETE FROM invoices WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
