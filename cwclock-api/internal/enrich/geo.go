// Package enrich looks up extra info about a visitor IP via external geo/IP
// services, used by the visitor interceptor (once per first-seen IP). It tries
// ip-api.com first, then ipinfo.io as a silent fallback, and silently returns no
// info if both fail — enrichment is best-effort and must never affect the request.
// Ported from the uprodit/sf-api implementation.
package enrich

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"
)

// Geo calls the external IP-geolocation providers.
type Geo struct {
	ipinfoToken string // optional; ipinfo works tokenless at low volume
	http        *http.Client
}

// NewGeo builds a Geo enricher. ipinfoToken may be empty.
func NewGeo(ipinfoToken string) *Geo {
	return &Geo{ipinfoToken: ipinfoToken, http: &http.Client{Timeout: 5 * time.Second}}
}

// Enrich returns geo info for ip, or (nil, nil) when both providers fail — the
// error is deliberately never surfaced (silent fail). Signature matches
// middleware.Enricher.
func (g *Geo) Enrich(ctx context.Context, ip string) (map[string]string, error) {
	if info := g.ipapi(ctx, ip); info != nil {
		return info, nil
	}
	if info := g.ipinfo(ctx, ip); info != nil {
		return info, nil
	}
	return nil, nil
}

// ipapi queries ip-api.com (free, no key, HTTP). Returns nil on any failure.
func (g *Geo) ipapi(ctx context.Context, ip string) map[string]string {
	url := "http://ip-api.com/json/" + ip +
		"?fields=status,country,countryCode,region,regionName,city,zip,lat,lon,timezone,isp,org,as,query"
	var r struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		RegionName  string  `json:"regionName"`
		City        string  `json:"city"`
		Zip         string  `json:"zip"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Timezone    string  `json:"timezone"`
		Isp         string  `json:"isp"`
		Org         string  `json:"org"`
		As          string  `json:"as"`
	}
	if !g.getJSON(ctx, url, &r) || r.Status != "success" {
		return nil
	}
	return nonEmpty(map[string]string{
		"source":      "ip-api.com",
		"country":     r.Country,
		"countryCode": r.CountryCode,
		"region":      r.RegionName,
		"city":        r.City,
		"zip":         r.Zip,
		"timezone":    r.Timezone,
		"isp":         r.Isp,
		"org":         r.Org,
		"asn":         r.As,
		"lat":         trimFloat(r.Lat),
		"lon":         trimFloat(r.Lon),
	})
}

// ipinfo queries ipinfo.io (HTTPS; optional token). Returns nil on any failure.
func (g *Geo) ipinfo(ctx context.Context, ip string) map[string]string {
	url := "https://ipinfo.io/" + ip + "/json"
	if g.ipinfoToken != "" {
		url += "?token=" + g.ipinfoToken
	}
	var r struct {
		City     string `json:"city"`
		Region   string `json:"region"`
		Country  string `json:"country"`
		Loc      string `json:"loc"`
		Org      string `json:"org"`
		Postal   string `json:"postal"`
		Timezone string `json:"timezone"`
		Bogon    bool   `json:"bogon"`
	}
	if !g.getJSON(ctx, url, &r) || r.Bogon || (r.Country == "" && r.City == "") {
		return nil
	}
	return nonEmpty(map[string]string{
		"source":      "ipinfo.io",
		"country":     r.Country,
		"countryCode": r.Country,
		"region":      r.Region,
		"city":        r.City,
		"zip":         r.Postal,
		"timezone":    r.Timezone,
		"org":         r.Org,
		"loc":         r.Loc,
	})
}

// getJSON GETs url and decodes JSON into v; false on any error or non-2xx.
func (g *Geo) getJSON(ctx context.Context, url string, v any) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}
	req.Header.Set("Accept", "application/json")
	resp, err := g.http.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<16))
	if err != nil {
		return false
	}
	return json.Unmarshal(body, v) == nil
}

func trimFloat(f float64) string {
	if f == 0 {
		return ""
	}
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func nonEmpty(m map[string]string) map[string]string {
	for k, v := range m {
		if v == "" {
			delete(m, k)
		}
	}
	return m
}
