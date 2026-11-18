// geoLabel renders an address's resolved location as "City, Region, Country (CC)"
// from the free-form geo map the API attaches to a visitor IP (see
// internal/enrich). Empty string when nothing resolved. networkLabel picks the
// most specific network name (ISP, else org, else ASN). Shared by the IP
// management screen and the per-user IPs modal.
export const geoLabel = (geo) => {
  if (!geo) return "";
  const label = [geo.city, geo.region, geo.country].filter(Boolean).join(", ");
  if (geo.countryCode) return label ? `${label} (${geo.countryCode})` : geo.countryCode;
  return label;
};

export const networkLabel = (geo) => (geo && (geo.isp || geo.org || geo.asn)) || "";
