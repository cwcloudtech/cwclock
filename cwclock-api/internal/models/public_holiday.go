package models

// HolidayType is how a non-working-day rule in the decision table resolves.
type HolidayType string

const (
	// HolidayFixed is a fixed civil date every year (Month/Day), e.g. 14 July.
	HolidayFixed HolidayType = "fixed"
	// HolidayEaster is a movable feast at EasterOffset days from Easter Sunday
	// (e.g. +1 = Easter Monday, +39 = Ascension, +50 = Whit Monday). Easter is
	// computed dynamically per year, so the date is correct every year.
	HolidayEaster HolidayType = "easter"
	// HolidayWeekend marks a whole weekday as non-working every week (the
	// weekend). It is configurable per country because the weekend isn't
	// Saturday/Sunday everywhere (e.g. Friday/Saturday in parts of the world).
	HolidayWeekend HolidayType = "weekend"
	// HolidayNthWeekday is the Nth occurrence of a weekday in a month (e.g. the
	// 4th Thursday of November = Thanksgiving, the last Monday of May = Memorial
	// Day). Common in the US and Australia.
	HolidayNthWeekday HolidayType = "nthWeekday"
)

// Weekdays are the valid values of a weekday rule's Weekday field (lowercase
// English), mapped to time.Weekday in the calendar.
var Weekdays = []string{"sunday", "monday", "tuesday", "wednesday", "thursday", "friday", "saturday"}

// PublicHoliday is one rule in the country-keyed non-working-day decision
// table: a holiday (fixed or easter-relative) or a non-working weekday. A
// country "activates" the movable feasts simply by having HolidayEaster rows,
// and sets its own weekend with HolidayWeekday rows.
type PublicHoliday struct {
	ID      string      `json:"id"`
	Country string      `json:"country"` // ISO 3166-1 alpha-2, uppercase
	Name    string      `json:"name"`
	Type    HolidayType `json:"type"`
	// Fixed rule: the civil-calendar month (1-12) and day (1-31).
	Month int `json:"month,omitempty"`
	Day   int `json:"day,omitempty"`
	// Easter rule: days offset from Easter Sunday (may be negative).
	EasterOffset int `json:"easterOffset,omitempty"`
	// Weekday rule (weekend) OR nth-weekday rule: the weekday, lowercase
	// English ("saturday", "monday").
	Weekday string `json:"weekday,omitempty"`
	// Nth-weekday rule: which occurrence of Weekday in Month — 1..5, or -1 for
	// the last (e.g. Month 11 / Thursday / 4 = Thanksgiving; Month 5 / Monday /
	// -1 = Memorial Day).
	Nth int `json:"nth,omitempty"`
}
