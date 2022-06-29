package demo

import (
	"fmt"
	"net"

	"github.com/IncSW/geoip2"
)

// DefaultDBPath default GeoIP2 database path.
const DefaultDBPath = "GeoLite2-ASN.mmdb"
const DefaultHeader = "x-real-ip"

// GeoIPResult GeoIPResult.
type GeoIPResult struct {
	asn uint32
	aso string
}

// LookupGeoIP2 LookupGeoIP2.
type LookupGeoIP2 func(ip net.IP) (*GeoIPResult, error)

// CreateASNDBLookup CreateASNDBLookup.
func CreateASNDBLookup(rdr *geoip2.ASNReader) LookupGeoIP2 {
	return func(ip net.IP) (*GeoIPResult, error) {
		rec, err := rdr.Lookup(ip)
		if err != nil {
			return nil, fmt.Errorf("%w", err)
		}
		retval := GeoIPResult{
			asn: rec.AutonomousSystemNumber,
			aso: rec.AutonomousSystemOrganization,
		}
		return &retval, nil
	}
}
