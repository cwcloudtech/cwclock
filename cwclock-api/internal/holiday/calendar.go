// Package holiday resolves a country's public-holiday rules (the DB decision
// table, keyed by ISO country code) into concrete dates, and tells working
// days from weekends and holidays over a range. The movable feasts are
// computed from Easter dynamically, so a rule set is correct for any year.
package holiday

import (
	"time"

	"cwclock-api/internal/models"
)

// Easter returns Easter Sunday (Gregorian) for year, by the Anonymous
// Gregorian algorithm (Meeus/Jones/Butcher). It is the anchor for the movable
// feasts (Easter Monday, Ascension, Whit Monday).
func Easter(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := ((19 * a) + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + (2 * e) + (2 * i) - h - k) % 7
	m := (a + (11 * h) + (22 * l)) / 451
	month := (h + l - (7 * m) + 114) / 31
	day := ((h + l - (7 * m) + 114) % 31) + 1
	return time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
}

// Calendar resolves one country's holiday rules into dates.
type Calendar struct {
	rules []models.PublicHoliday
}

// NewCalendar builds a calendar from a country's rules (as returned by the
// holiday store). An empty rule set means "no configured holidays": only
// weekends are non-working.
func NewCalendar(rules []models.PublicHoliday) *Calendar {
	return &Calendar{rules: rules}
}

// dateFor resolves one dated rule (fixed or easter) to its date in the given
// year, or ok=false when the rule is malformed or not a dated rule.
func (c *Calendar) dateFor(rule models.PublicHoliday, year int) (time.Time, bool) {
	switch rule.Type {
	case models.HolidayFixed:
		if rule.Month < 1 || rule.Month > 12 || rule.Day < 1 || rule.Day > 31 {
			return time.Time{}, false
		}
		return time.Date(year, time.Month(rule.Month), rule.Day, 0, 0, 0, 0, time.UTC), true
	case models.HolidayEaster:
		return Easter(year).AddDate(0, 0, rule.EasterOffset), true
	}
	return time.Time{}, false
}

// weekdayNumbers maps lowercase English weekday names to time.Weekday.
var weekdayNumbers = map[string]time.Weekday{
	"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday,
	"wednesday": time.Wednesday, "thursday": time.Thursday, "friday": time.Friday,
	"saturday": time.Saturday,
}

// ParseWeekday maps a lowercase English weekday name to time.Weekday.
func ParseWeekday(name string) (time.Weekday, bool) {
	wd, ok := weekdayNumbers[name]
	return wd, ok
}

// NonWorkingWeekdays is the set of weekdays this country marks non-working (its
// weekend), from its HolidayWeekday rules. Empty when the country configured
// none — the caller then falls back to Saturday/Sunday.
func (c *Calendar) NonWorkingWeekdays() map[time.Weekday]bool {
	out := map[time.Weekday]bool{}
	for _, rule := range c.rules {
		if rule.Type == models.HolidayWeekend {
			if wd, ok := weekdayNumbers[rule.Weekday]; ok {
				out[wd] = true
			}
		}
	}
	return out
}

// Holidays returns dayKey("YYYY-MM-DD") -> holiday name for every holiday that
// falls within [from, to] inclusive. Both bounds are interpreted as whole days
// in UTC.
func (c *Calendar) Holidays(from, to time.Time) map[string]string {
	out := map[string]string{}
	if to.Before(from) {
		return out
	}
	for year := from.Year(); year <= to.Year(); year++ {
		for _, rule := range c.rules {
			d, ok := c.dateFor(rule, year)
			if !ok {
				continue
			}
			if d.Before(from) || d.After(to) {
				continue
			}
			out[DayKey(d)] = rule.Name
		}
	}
	return out
}

// DayKey formats a date as the "YYYY-MM-DD" key used for time entries' day.
func DayKey(d time.Time) string {
	return d.Format("2006-01-02")
}

// IsWeekend reports whether d falls on a Saturday or Sunday.
func IsWeekend(d time.Time) bool {
	wd := d.Weekday()
	return wd == time.Saturday || wd == time.Sunday
}
