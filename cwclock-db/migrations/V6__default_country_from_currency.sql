-- Backfills organizations/clients that have no country set yet, inferring
-- one from their currency (e.g. EUR -> France) rather than leaving them
-- unable to satisfy the now-required country field (ai-instruct-35).
-- Clients have no currency of their own, so they inherit their parent
-- organization's currency for this mapping.

-- CAD/HKD/SGD are already in the currencies table (see V4/V5) but had no
-- matching country row, so the mapping below would have nowhere to point
-- them - same gap as TRY/Turkey in V5, filled here for the same reason.
INSERT INTO countries (iso_code, name, currency_iso_code, sort_order) VALUES
    ('CA', 'Canada', 'CAD', 4000),
    ('HK', 'Hong Kong', 'HKD', 15000),
    ('SG', 'Singapore', 'SGD', 16000)
ON CONFLICT (iso_code) DO NOTHING;

UPDATE organizations o
SET data = jsonb_set(o.data, '{country}', to_jsonb(mapping.country_iso_code))
FROM (VALUES
    ('EUR', 'FR'), ('USD', 'US'), ('GBP', 'GB'), ('CAD', 'CA'), ('CHF', 'CH'),
    ('TND', 'TN'), ('DZD', 'DZ'), ('MAD', 'MA'), ('TRY', 'TR'), ('EGP', 'EG'),
    ('SAR', 'SA'), ('AED', 'AE'), ('QAR', 'QA'), ('CNY', 'CN'), ('HKD', 'HK'),
    ('SGD', 'SG'), ('JPY', 'JP'), ('AUD', 'AU'), ('NZD', 'NZ'), ('KWD', 'KW')
) AS mapping(currency_iso_code, country_iso_code)
WHERE o.data->>'currency' = mapping.currency_iso_code
  AND (o.data->>'country' IS NULL OR trim(o.data->>'country') = '');

UPDATE clients cl
SET data = jsonb_set(cl.data, '{country}', to_jsonb(mapping.country_iso_code))
FROM organizations o, (VALUES
    ('EUR', 'FR'), ('USD', 'US'), ('GBP', 'GB'), ('CAD', 'CA'), ('CHF', 'CH'),
    ('TND', 'TN'), ('DZD', 'DZ'), ('MAD', 'MA'), ('TRY', 'TR'), ('EGP', 'EG'),
    ('SAR', 'SA'), ('AED', 'AE'), ('QAR', 'QA'), ('CNY', 'CN'), ('HKD', 'HK'),
    ('SGD', 'SG'), ('JPY', 'JP'), ('AUD', 'AU'), ('NZD', 'NZ'), ('KWD', 'KW')
) AS mapping(currency_iso_code, country_iso_code)
WHERE cl.organization_id = o.id
  AND o.data->>'currency' = mapping.currency_iso_code
  AND (cl.data->>'country' IS NULL OR trim(cl.data->>'country') = '');
