-- Countries, currencies and per-country identification fields, replacing
-- the CWCLOCK_ALLOWED_CURRENCIES env var and free-text country inputs with
-- real reference tables (see ai-instruct-35).

CREATE TABLE currencies (
    iso_code VARCHAR(3) PRIMARY KEY,
    name VARCHAR(255) NOT NULL
);

CREATE TABLE countries (
    iso_code VARCHAR(2) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    currency_iso_code VARCHAR(3) NOT NULL REFERENCES currencies(iso_code)
);

CREATE INDEX idx_countries_currency_iso_code ON countries(currency_iso_code);

-- Per-country business identification fields to display (ai-instruct-35's
-- decision table): e.g. FR gets SIRET/SIREN/NAF, TN gets MF, every EU
-- country gets a VAT Code row. A country with no rows here falls back to
-- the generic "Identification number" field at the API layer.
CREATE TABLE fields (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    iso_code VARCHAR(2) NOT NULL REFERENCES countries(iso_code),
    name VARCHAR(255) NOT NULL
);

CREATE INDEX idx_fields_iso_code ON fields(iso_code);

INSERT INTO currencies (iso_code, name) VALUES
    ('AED', 'UAE Dirham'),
    ('AUD', 'Australian Dollar'),
    ('CAD', 'Canadian Dollar'),
    ('CHF', 'Swiss Franc'),
    ('CNY', 'Chinese Yuan'),
    ('DZD', 'Algerian Dinar'),
    ('EGP', 'Egyptian Pound'),
    ('EUR', 'Euro'),
    ('GBP', 'British Pound'),
    ('HKD', 'Hong Kong Dollar'),
    ('JPY', 'Japanese Yen'),
    ('KWD', 'Kuwaiti dinar'),
    ('MAD', 'Moroccan Dirham'),
    ('NZD', 'New Zealand Dollar'),
    ('QAR', 'Qatari Riyal'),
    ('SAR', 'Saudi Riyal'),
    ('SGD', 'Singapore Dollar'),
    ('TND', 'Tunisian Dinar'),
    ('USD', 'US Dollar')
ON CONFLICT (iso_code) DO NOTHING;

INSERT INTO countries (iso_code, name, currency_iso_code) VALUES
    ('AT', 'Austria', 'EUR'),
    ('BE', 'Belgium', 'EUR'),
    ('HR', 'Croatia', 'EUR'),
    ('CY', 'Cyprus', 'EUR'),
    ('EE', 'Estonia', 'EUR'),
    ('FI', 'Finland', 'EUR'),
    ('FR', 'France', 'EUR'),
    ('DE', 'Germany', 'EUR'),
    ('GR', 'Greece', 'EUR'),
    ('IE', 'Ireland', 'EUR'),
    ('IT', 'Italy', 'EUR'),
    ('LV', 'Latvia', 'EUR'),
    ('LT', 'Lithuania', 'EUR'),
    ('LU', 'Luxembourg', 'EUR'),
    ('MT', 'Malta', 'EUR'),
    ('NL', 'Netherlands', 'EUR'),
    ('PT', 'Portugal', 'EUR'),
    ('SK', 'Slovakia', 'EUR'),
    ('SI', 'Slovenia', 'EUR'),
    ('ES', 'Spain', 'EUR'),
    ('GB', 'United Kingdom', 'GBP'),
    ('CH', 'Switzerland', 'CHF'),
    ('LI', 'Liechtenstein', 'CHF'),
    ('MC', 'Monaco', 'EUR'),
    ('SM', 'San Marino', 'EUR'),
    ('VA', 'Vatican City', 'EUR'),
    ('AD', 'Andorra', 'EUR'),
    ('ME', 'Montenegro', 'EUR'),
    ('US', 'United States', 'USD'),
    ('TN', 'Tunisia', 'TND'),
    ('DZ', 'Algeria', 'DZD'),
    ('MA', 'Morocco', 'MAD'),
    ('EG', 'Egypt', 'EGP'),
    ('SA', 'Saudi Arabia', 'SAR'),
    ('AE', 'United Arab Emirates', 'AED'),
    ('QA', 'Qatar', 'QAR'),
    ('KW', 'Kuwait', 'KWD'),
    ('CN', 'China', 'CNY'),
    ('JP', 'Japan', 'JPY'),
    ('AU', 'Australia', 'AUD'),
    ('NZ', 'New Zealand', 'NZD')
ON CONFLICT (iso_code) DO NOTHING;

INSERT INTO fields (iso_code, name) VALUES
    ('FR', 'SIRET'),
    ('FR', 'SIREN'),
    ('FR', 'NAF'),
    ('TN', 'MF'),
    ('AT', 'VAT Code'),
    ('BE', 'VAT Code'),
    ('HR', 'VAT Code'),
    ('CY', 'VAT Code'),
    ('EE', 'VAT Code'),
    ('FI', 'VAT Code'),
    ('FR', 'VAT Code'),
    ('DE', 'VAT Code'),
    ('GR', 'VAT Code'),
    ('IE', 'VAT Code'),
    ('IT', 'VAT Code'),
    ('LV', 'VAT Code'),
    ('LT', 'VAT Code'),
    ('LU', 'VAT Code'),
    ('MT', 'VAT Code'),
    ('NL', 'VAT Code'),
    ('PT', 'VAT Code'),
    ('SK', 'VAT Code'),
    ('SI', 'VAT Code'),
    ('ES', 'VAT Code');

-- Backfills existing free-text country values (organizations/clients store
-- it inside their JSONB data blob) to the matching ISO code, e.g. "France",
-- "france" and "FRANCE" all become "FR". Idempotent: once a row's country
-- is already an ISO code, it no longer matches any country *name* here, so
-- re-running this UPDATE is a no-op for it.
UPDATE organizations o
SET data = jsonb_set(o.data, '{country}', to_jsonb(c.iso_code))
FROM countries c
WHERE lower(trim(o.data->>'country')) = lower(c.name);

UPDATE clients cl
SET data = jsonb_set(cl.data, '{country}', to_jsonb(c.iso_code))
FROM countries c
WHERE lower(trim(cl.data->>'country')) = lower(c.name);
