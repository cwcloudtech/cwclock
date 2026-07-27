CREATE TABLE invoices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    client_id UUID NOT NULL REFERENCES clients(id) ON DELETE CASCADE,
    data JSONB NOT NULL DEFAULT '{}'::jsonb,
    pdf BYTEA NOT NULL,
    selected_begin_date DATE NOT NULL,
    selected_end_date DATE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_invoices_organization_id ON invoices(organization_id);
CREATE INDEX idx_invoices_client_id ON invoices(client_id);

-- Enforces the "{CLIENT_NAME_CAPITALIZED}{YYYYMMDD}{incremental_number}"
-- invoice numbering scheme's uniqueness per organization, so concurrent
-- invoice generation can't ever produce two invoices sharing a number.
CREATE UNIQUE INDEX idx_invoices_org_number ON invoices(organization_id, (data->>'number'));
