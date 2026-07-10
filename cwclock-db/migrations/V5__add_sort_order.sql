-- Adds a manually-curated sort order to currencies and countries
-- (ai-instruct-36), replacing the implicit alphabetical ordering used since
-- ai-instruct-35. Values are spaced by 1000 so a new entry can be inserted
-- between two existing ones (e.g. 1500 between EUR and USD) without
-- renumbering the rest. Rows with no explicit rank (added below or later)
-- default to 999999, so they sort after every ranked entry.
ALTER TABLE currencies ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 999999;
ALTER TABLE countries ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 999999;

-- TRY/Turkey is part of the ranked list below but was missing from the
-- V4 seed data; added here so it has a currency/country row to rank.
INSERT INTO currencies (iso_code, name) VALUES ('TRY', 'Turkish Lira')
ON CONFLICT (iso_code) DO NOTHING;

INSERT INTO countries (iso_code, name, currency_iso_code) VALUES ('TR', 'Turkiye', 'TRY')
ON CONFLICT (iso_code) DO NOTHING;

UPDATE currencies SET sort_order = ranked.sort_order
FROM (VALUES
    ('EUR', 1000), ('USD', 2000), ('GBP', 3000), ('CAD', 4000), ('CHF', 5000),
    ('TND', 6000), ('DZD', 7000), ('MAD', 8000), ('TRY', 9000), ('EGP', 10000),
    ('SAR', 11000), ('AED', 12000), ('QAR', 13000), ('CNY', 14000), ('HKD', 15000),
    ('SGD', 16000), ('JPY', 17000), ('AUD', 18000), ('NZD', 19000)
) AS ranked(iso_code, sort_order)
WHERE currencies.iso_code = ranked.iso_code;

UPDATE countries SET sort_order = ranked.sort_order
FROM (VALUES
    ('EUR', 1000), ('USD', 2000), ('GBP', 3000), ('CAD', 4000), ('CHF', 5000),
    ('TND', 6000), ('DZD', 7000), ('MAD', 8000), ('TRY', 9000), ('EGP', 10000),
    ('SAR', 11000), ('AED', 12000), ('QAR', 13000), ('CNY', 14000), ('HKD', 15000),
    ('SGD', 16000), ('JPY', 17000), ('AUD', 18000), ('NZD', 19000)
) AS ranked(currency_iso_code, sort_order)
WHERE countries.currency_iso_code = ranked.currency_iso_code;
