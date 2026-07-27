CREATE TABLE mail_counter (
    orga_id UUID PRIMARY KEY REFERENCES organizations(id) ON DELETE CASCADE,
    count INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
