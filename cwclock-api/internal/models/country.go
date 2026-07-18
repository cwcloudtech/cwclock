package models

// Country is an ISO 3166-1 alpha-2 country, with the currency organizations
// in it are billed in by default (see Currency).
type Country struct {
	ISO      string `json:"iso"`
	Name     string `json:"name"`
	Currency string `json:"currency"`
}

// Currency is an ISO 4217 currency organizations may be billed in.
type Currency struct {
	ISO  string `json:"iso"`
	Name string `json:"name"`
}

// FallbackCurrency is applied to an organization when neither an explicit
// currency nor a resolvable country default is available.
const FallbackCurrency = "EUR"

// identificationFieldFallback is returned by the fields decision table for
// a country with no fields rows of its own (ai-instruct-35: "For the rest,
// by default put 'Identification number'").
const identificationFieldFallback = "Identification number"

// vatCodeField is the well-known field name every EU country's fields rows
// use; ResolveFields uses it to tell "this country already has its own
// identification field(s)" apart from "it only has the EU VAT Code row".
const vatCodeField = "VAT Code"

// frenchSpeakingCountries is the client_language(country_iso,
// language_iso_code) decision table (ai-instruct-51): the closed set of
// country ISO codes whose invoice emails go out in French. Any country not
// listed here keeps English.
var frenchSpeakingCountries = map[string]bool{
	"FR": true, // France
	"BE": true, // Belgium
	"TN": true, // Tunisia
	"DZ": true, // Algeria
	"MA": true, // Morocco
}

// ClientLanguage is the client_language(country_iso, language_iso_code)
// decision table: the ISO 639-1 language code invoice emails should be sent
// in for a client based in countryISO.
func ClientLanguage(countryISO string) string {
	if frenchSpeakingCountries[countryISO] {
		return "fr"
	}
	return "en"
}

// ResolveFields applies the "For the rest, by default put 'Identification
// number'" fallback on top of a country's raw fields rows: a country with
// no rows at all gets just the fallback, and a country whose only row is
// the EU VAT Code gets the fallback prepended to it. A country with its own
// identification field(s) (e.g. FR's SIRET/SIREN/NAF, TN's MF) is returned
// as-is.
func ResolveFields(rows []string) []string {
	for _, name := range rows {
		if name != vatCodeField {
			return rows
		}
	}
	return append([]string{identificationFieldFallback}, rows...)
}
