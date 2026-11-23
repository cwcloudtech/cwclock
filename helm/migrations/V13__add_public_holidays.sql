-- Non-working-day decision table, keyed by ISO 3166-1 alpha-2 country code.
-- Drives which days count as non-working when filling a range of working days
-- for a client/project. Each row is one RULE:
--   - type "fixed"   : a civil-calendar date, {month, day}
--   - type "easter"  : a movable feast computed from Easter Sunday, {easterOffset}
--     (Easter is computed dynamically in code, so it is correct every year; a
--     country "activates" the movable feasts simply by having these rows).
--   - type "weekend" : a whole weekday that is non-working every week, {weekday}
--     (lowercase English). The weekend is configurable per country because it
--     isn't Saturday/Sunday everywhere (e.g. Friday/Saturday in some countries).
--     A country with NO weekend rows falls back to Saturday/Sunday in code, so
--     an unconfigured country still never fills its weekends.
-- Resolution at fill time is client country first, then the organization's.
-- Only France (FR) and Belgium (BE) are seeded; every other country has no
-- configured rules (and falls back to the Saturday/Sunday weekend).
CREATE TABLE public_holidays (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    country    TEXT NOT NULL,
    data       JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_public_holidays_country ON public_holidays(country);

-- France — Saturday/Sunday weekend + 11 national public holidays.
INSERT INTO public_holidays (country, data) VALUES
    ('FR', '{"name":"Samedi","type":"weekend","weekday":"saturday"}'::jsonb),
    ('FR', '{"name":"Dimanche","type":"weekend","weekday":"sunday"}'::jsonb),
    ('FR', '{"name":"Jour de l''an","type":"fixed","month":1,"day":1}'::jsonb),
    ('FR', '{"name":"Lundi de Pâques","type":"easter","easterOffset":1}'::jsonb),
    ('FR', '{"name":"Fête du travail","type":"fixed","month":5,"day":1}'::jsonb),
    ('FR', '{"name":"Victoire 1945","type":"fixed","month":5,"day":8}'::jsonb),
    ('FR', '{"name":"Ascension","type":"easter","easterOffset":39}'::jsonb),
    ('FR', '{"name":"Lundi de Pentecôte","type":"easter","easterOffset":50}'::jsonb),
    ('FR', '{"name":"Fête nationale","type":"fixed","month":7,"day":14}'::jsonb),
    ('FR', '{"name":"Assomption","type":"fixed","month":8,"day":15}'::jsonb),
    ('FR', '{"name":"Toussaint","type":"fixed","month":11,"day":1}'::jsonb),
    ('FR', '{"name":"Armistice 1918","type":"fixed","month":11,"day":11}'::jsonb),
    ('FR', '{"name":"Noël","type":"fixed","month":12,"day":25}'::jsonb);

-- Belgium — Saturday/Sunday weekend + 10 legal public holidays. Same as France
-- except: no "Victoire 1945" (8 May), and the national day is 21 July (not 14).
INSERT INTO public_holidays (country, data) VALUES
    ('BE', '{"name":"Samedi","type":"weekend","weekday":"saturday"}'::jsonb),
    ('BE', '{"name":"Dimanche","type":"weekend","weekday":"sunday"}'::jsonb),
    ('BE', '{"name":"Jour de l''an","type":"fixed","month":1,"day":1}'::jsonb),
    ('BE', '{"name":"Lundi de Pâques","type":"easter","easterOffset":1}'::jsonb),
    ('BE', '{"name":"Fête du travail","type":"fixed","month":5,"day":1}'::jsonb),
    ('BE', '{"name":"Ascension","type":"easter","easterOffset":39}'::jsonb),
    ('BE', '{"name":"Lundi de Pentecôte","type":"easter","easterOffset":50}'::jsonb),
    ('BE', '{"name":"Fête nationale","type":"fixed","month":7,"day":21}'::jsonb),
    ('BE', '{"name":"Assomption","type":"fixed","month":8,"day":15}'::jsonb),
    ('BE', '{"name":"Toussaint","type":"fixed","month":11,"day":1}'::jsonb),
    ('BE', '{"name":"Armistice","type":"fixed","month":11,"day":11}'::jsonb),
    ('BE', '{"name":"Noël","type":"fixed","month":12,"day":25}'::jsonb);
