-- More countries for the non-working-day decision table (see V13 for the rule
-- shapes): US, Australia, UK, Germany, Tunisia, Algeria, Egypt, Saudi Arabia,
-- UAE. Only the Gregorian-computable holidays are seeded — the fixed civil
-- dates, the easter-relative feasts, and the "nth weekday of a month" ones
-- (type "nthWeekday": {month, weekday, nth}; nth 1..5, or -1 for the last).
--
-- Islamic (Hijri/lunar) holidays — Eid al-Fitr, Eid al-Adha, Islamic New Year,
-- Mawlid — and Egypt's Coptic Sham El-Nessim are NOT seeded: they can't be
-- computed from these rule types, so an admin adds them per year from the
-- "Jours fériés" screen. Weekends: Friday+Saturday for DZ/EG/SA/AE, Saturday+
-- Sunday for US/AU/GB/DE/TN (matching FR/BE in V13).

-- United States (Sat/Sun) — 11 federal holidays.
INSERT INTO public_holidays (country, data) VALUES
    ('US', '{"name":"Saturday","type":"weekend","weekday":"saturday"}'::jsonb),
    ('US', '{"name":"Sunday","type":"weekend","weekday":"sunday"}'::jsonb),
    ('US', '{"name":"New Year''s Day","type":"fixed","month":1,"day":1}'::jsonb),
    ('US', '{"name":"Martin Luther King Jr. Day","type":"nthWeekday","month":1,"weekday":"monday","nth":3}'::jsonb),
    ('US', '{"name":"Presidents'' Day","type":"nthWeekday","month":2,"weekday":"monday","nth":3}'::jsonb),
    ('US', '{"name":"Memorial Day","type":"nthWeekday","month":5,"weekday":"monday","nth":-1}'::jsonb),
    ('US', '{"name":"Juneteenth","type":"fixed","month":6,"day":19}'::jsonb),
    ('US', '{"name":"Independence Day","type":"fixed","month":7,"day":4}'::jsonb),
    ('US', '{"name":"Labor Day","type":"nthWeekday","month":9,"weekday":"monday","nth":1}'::jsonb),
    ('US', '{"name":"Columbus Day","type":"nthWeekday","month":10,"weekday":"monday","nth":2}'::jsonb),
    ('US', '{"name":"Veterans Day","type":"fixed","month":11,"day":11}'::jsonb),
    ('US', '{"name":"Thanksgiving Day","type":"nthWeekday","month":11,"weekday":"thursday","nth":4}'::jsonb),
    ('US', '{"name":"Christmas Day","type":"fixed","month":12,"day":25}'::jsonb);

-- Australia (Sat/Sun) — national set; some holidays (King's Birthday, Labour
-- Day) vary by state, so adjust per state if needed.
INSERT INTO public_holidays (country, data) VALUES
    ('AU', '{"name":"Saturday","type":"weekend","weekday":"saturday"}'::jsonb),
    ('AU', '{"name":"Sunday","type":"weekend","weekday":"sunday"}'::jsonb),
    ('AU', '{"name":"New Year''s Day","type":"fixed","month":1,"day":1}'::jsonb),
    ('AU', '{"name":"Australia Day","type":"fixed","month":1,"day":26}'::jsonb),
    ('AU', '{"name":"Good Friday","type":"easter","easterOffset":-2}'::jsonb),
    ('AU', '{"name":"Easter Monday","type":"easter","easterOffset":1}'::jsonb),
    ('AU', '{"name":"Anzac Day","type":"fixed","month":4,"day":25}'::jsonb),
    ('AU', '{"name":"King''s Birthday","type":"nthWeekday","month":6,"weekday":"monday","nth":2}'::jsonb),
    ('AU', '{"name":"Christmas Day","type":"fixed","month":12,"day":25}'::jsonb),
    ('AU', '{"name":"Boxing Day","type":"fixed","month":12,"day":26}'::jsonb);

-- United Kingdom (Sat/Sun) — England & Wales bank holidays (Scotland / NI
-- differ).
INSERT INTO public_holidays (country, data) VALUES
    ('GB', '{"name":"Saturday","type":"weekend","weekday":"saturday"}'::jsonb),
    ('GB', '{"name":"Sunday","type":"weekend","weekday":"sunday"}'::jsonb),
    ('GB', '{"name":"New Year''s Day","type":"fixed","month":1,"day":1}'::jsonb),
    ('GB', '{"name":"Good Friday","type":"easter","easterOffset":-2}'::jsonb),
    ('GB', '{"name":"Easter Monday","type":"easter","easterOffset":1}'::jsonb),
    ('GB', '{"name":"Early May Bank Holiday","type":"nthWeekday","month":5,"weekday":"monday","nth":1}'::jsonb),
    ('GB', '{"name":"Spring Bank Holiday","type":"nthWeekday","month":5,"weekday":"monday","nth":-1}'::jsonb),
    ('GB', '{"name":"Summer Bank Holiday","type":"nthWeekday","month":8,"weekday":"monday","nth":-1}'::jsonb),
    ('GB', '{"name":"Christmas Day","type":"fixed","month":12,"day":25}'::jsonb),
    ('GB', '{"name":"Boxing Day","type":"fixed","month":12,"day":26}'::jsonb);

-- Germany (Sat/Sun) — the nationwide federal holidays (several states add more).
INSERT INTO public_holidays (country, data) VALUES
    ('DE', '{"name":"Saturday","type":"weekend","weekday":"saturday"}'::jsonb),
    ('DE', '{"name":"Sunday","type":"weekend","weekday":"sunday"}'::jsonb),
    ('DE', '{"name":"New Year''s Day","type":"fixed","month":1,"day":1}'::jsonb),
    ('DE', '{"name":"Good Friday","type":"easter","easterOffset":-2}'::jsonb),
    ('DE', '{"name":"Easter Monday","type":"easter","easterOffset":1}'::jsonb),
    ('DE', '{"name":"Labour Day","type":"fixed","month":5,"day":1}'::jsonb),
    ('DE', '{"name":"Ascension Day","type":"easter","easterOffset":39}'::jsonb),
    ('DE', '{"name":"Whit Monday","type":"easter","easterOffset":50}'::jsonb),
    ('DE', '{"name":"German Unity Day","type":"fixed","month":10,"day":3}'::jsonb),
    ('DE', '{"name":"Christmas Day","type":"fixed","month":12,"day":25}'::jsonb),
    ('DE', '{"name":"Boxing Day","type":"fixed","month":12,"day":26}'::jsonb);

-- Tunisia (Sat/Sun) — secular public holidays (Islamic ones added per year).
INSERT INTO public_holidays (country, data) VALUES
    ('TN', '{"name":"Saturday","type":"weekend","weekday":"saturday"}'::jsonb),
    ('TN', '{"name":"Sunday","type":"weekend","weekday":"sunday"}'::jsonb),
    ('TN', '{"name":"New Year''s Day","type":"fixed","month":1,"day":1}'::jsonb),
    ('TN', '{"name":"Revolution and Youth Day","type":"fixed","month":1,"day":14}'::jsonb),
    ('TN', '{"name":"Independence Day","type":"fixed","month":3,"day":20}'::jsonb),
    ('TN', '{"name":"Martyrs'' Day","type":"fixed","month":4,"day":9}'::jsonb),
    ('TN', '{"name":"Labour Day","type":"fixed","month":5,"day":1}'::jsonb),
    ('TN', '{"name":"Republic Day","type":"fixed","month":7,"day":25}'::jsonb),
    ('TN', '{"name":"Women''s Day","type":"fixed","month":8,"day":13}'::jsonb),
    ('TN', '{"name":"Evacuation Day","type":"fixed","month":10,"day":15}'::jsonb);

-- Algeria (Fri/Sat) — secular public holidays (Islamic ones added per year).
INSERT INTO public_holidays (country, data) VALUES
    ('DZ', '{"name":"Friday","type":"weekend","weekday":"friday"}'::jsonb),
    ('DZ', '{"name":"Saturday","type":"weekend","weekday":"saturday"}'::jsonb),
    ('DZ', '{"name":"New Year''s Day","type":"fixed","month":1,"day":1}'::jsonb),
    ('DZ', '{"name":"Yennayer (Amazigh New Year)","type":"fixed","month":1,"day":12}'::jsonb),
    ('DZ', '{"name":"Labour Day","type":"fixed","month":5,"day":1}'::jsonb),
    ('DZ', '{"name":"Independence Day","type":"fixed","month":7,"day":5}'::jsonb),
    ('DZ', '{"name":"Revolution Day","type":"fixed","month":11,"day":1}'::jsonb);

-- Egypt (Fri/Sat) — secular + Coptic Christmas (fixed 7 Jan); Sham El-Nessim
-- (Coptic Easter Monday) is not seeded (Julian calendar).
INSERT INTO public_holidays (country, data) VALUES
    ('EG', '{"name":"Friday","type":"weekend","weekday":"friday"}'::jsonb),
    ('EG', '{"name":"Saturday","type":"weekend","weekday":"saturday"}'::jsonb),
    ('EG', '{"name":"Coptic Christmas","type":"fixed","month":1,"day":7}'::jsonb),
    ('EG', '{"name":"Revolution Day (Police Day)","type":"fixed","month":1,"day":25}'::jsonb),
    ('EG', '{"name":"Sinai Liberation Day","type":"fixed","month":4,"day":25}'::jsonb),
    ('EG', '{"name":"Labour Day","type":"fixed","month":5,"day":1}'::jsonb),
    ('EG', '{"name":"Revolution Day","type":"fixed","month":7,"day":23}'::jsonb),
    ('EG', '{"name":"Armed Forces Day","type":"fixed","month":10,"day":6}'::jsonb);

-- Saudi Arabia (Fri/Sat) — the Gregorian national days; the Eids (Islamic) are
-- added per year.
INSERT INTO public_holidays (country, data) VALUES
    ('SA', '{"name":"Friday","type":"weekend","weekday":"friday"}'::jsonb),
    ('SA', '{"name":"Saturday","type":"weekend","weekday":"saturday"}'::jsonb),
    ('SA', '{"name":"Founding Day","type":"fixed","month":2,"day":22}'::jsonb),
    ('SA', '{"name":"National Day","type":"fixed","month":9,"day":23}'::jsonb);

-- United Arab Emirates (Fri/Sat, per request) — secular days; the Islamic
-- holidays are added per year.
INSERT INTO public_holidays (country, data) VALUES
    ('AE', '{"name":"Friday","type":"weekend","weekday":"friday"}'::jsonb),
    ('AE', '{"name":"Saturday","type":"weekend","weekday":"saturday"}'::jsonb),
    ('AE', '{"name":"New Year''s Day","type":"fixed","month":1,"day":1}'::jsonb),
    ('AE', '{"name":"Commemoration Day","type":"fixed","month":12,"day":1}'::jsonb),
    ('AE', '{"name":"National Day","type":"fixed","month":12,"day":2}'::jsonb),
    ('AE', '{"name":"National Day (second day)","type":"fixed","month":12,"day":3}'::jsonb);
