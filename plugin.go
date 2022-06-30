package demo

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"

	"github.com/IncSW/geoip2"
)

type Config struct {
	DatabaseFilePath string
	AllowedASNs      []string
	DisallowedASNs   []string
	Enabled          bool
	Header           string
}

func CreateConfig() *Config {
	return &Config{
		Enabled:          true,
		DatabaseFilePath: DefaultDBPath,
		Header:           DefaultHeader,
	}
}

// Plugin is the traefik ip2location plugin implementation.
type Plugin struct {
	next           http.Handler
	name           string
	allowedASNs    []string
	disallowedASNs []string
	enabled        bool
	lookup         LookupGeoIP2
	header         string
	dbPath         string
}

// New creates a new plugin handler.
func New(_ context.Context, next http.Handler, cfg *Config, name string) (http.Handler, error) {
	if len(cfg.DisallowedASNs) > 0 && len(cfg.AllowedASNs) > 0 {
		return nil, errors.New("either allowed asn or disallowed asn could be set at once")
	}

	if !cfg.Enabled {
		log.Printf("%s: disabled", name)

		return &Plugin{
			next: next,
			name: name,
		}, nil
	}
	log.Printf("[geoip2] DBPathL `%s`", cfg.DatabaseFilePath)
	if _, err := os.Stat(cfg.DatabaseFilePath); err != nil {
		log.Printf("[geoip2] DB `%s' not found: %v", cfg.DatabaseFilePath, err)
		return &Plugin{
			lookup: nil,
			next:   next,
			name:   name,
		}, nil
	}

	var lookup LookupGeoIP2
	rdr, err := geoip2.NewASNReaderFromFile(cfg.DatabaseFilePath)
	if err != nil {
		log.Printf("[geoip2] DB `%s' not initialized: %v", cfg.DatabaseFilePath, err)
	}
	lookup = CreateASNDBLookup(rdr)
	return &Plugin{
		next:           next,
		name:           name,
		enabled:        cfg.Enabled,
		allowedASNs:    cfg.AllowedASNs,
		disallowedASNs: cfg.DisallowedASNs,
		header:         cfg.Header,
		lookup:         lookup,
		dbPath:         cfg.DatabaseFilePath,
	}, nil
}

// ServeHTTP implements http.Handler interface.
func (p *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	p.header = "X-Real-Ip"
	p.enabled = true
	log.Println(p.allowedASNs)
	log.Println(p.disallowedASNs)
	if !p.enabled {
		p.next.ServeHTTP(rw, req)
		return
	}
	ips := p.GetRemoteIPs(req, p.header)
	for _, ip := range ips {
		ipDetail, err := p.CheckAllowed(ip)
		log.Printf("ASN: %d", ipDetail.asn)
		if err != nil {
			if errors.Is(err, ErrNotAllowed) {
				log.Printf("%s: %s - access denied for %s (%v)", p.name, req.Host, ip, ipDetail)
				rw.WriteHeader(http.StatusForbidden)

			} else {
				log.Printf("%s: %s - %v", p.name, req.Host, err)
				p.next.ServeHTTP(rw, req)
			}
		}
	}

	p.next.ServeHTTP(rw, req)

}

// GetRemoteIPs collects the remote IPs from the X-Forwarded-For and X-Real-IP headers.
func (p *Plugin) GetRemoteIPs(req *http.Request, header string) (ips []string) {
	ipMap := make(map[string]struct{})

	if ip := req.Header.Get(header); ip != "" {
		ipMap[ip] = struct{}{}
	}

	for ip := range ipMap {
		ips = append(ips, ip)
	}

	return
}

var ErrNotAllowed = errors.New("not allowed")

// CheckAllowed checks whether a given IP address is allowed according to the configured allowed ans.
func (p *Plugin) CheckAllowed(ip string) (*GeoIPResult, error) {
	asn, err := p.Lookup(ip)
	if err != nil {
		return &GeoIPResult{}, fmt.Errorf("lookup of %s failed: %w", ip, err)
	}
	if len(p.allowedASNs) > 0 {
		var allowed bool
		for _, allowedASN := range p.allowedASNs {
			if allowedASN == strconv.Itoa(int(asn.asn)) {
				allowed = true
				break
			}
		}
		if !allowed {
			return asn, ErrNotAllowed
		}
	}

	if len(p.disallowedASNs) > 0 {
		allowed := true
		for _, disallowed := range p.disallowedASNs {
			if disallowed == strconv.Itoa(int(asn.asn)) {
				allowed = false
				break
			}
		}
		if !allowed {
			return asn, ErrNotAllowed
		}
	}

	return asn, nil
}

// Lookup ASN from a given IP address.
func (p *Plugin) Lookup(ip string) (*GeoIPResult, error) {
	return p.lookup(net.ParseIP(ip))
}

type noopHandler struct{}

func (n noopHandler) ServeHTTP(rw http.ResponseWriter, _ *http.Request) {
	rw.WriteHeader(http.StatusTeapot)
}

func CreatePlugin(header string, ans []string, pluginName string) (http.Handler, error) {
	conf := CreateConfig()
	conf.Header = header
	plug, err := New(context.TODO(), &noopHandler{}, conf, pluginName)
	if err != nil {
		return nil, err
	}
	return plug, nil
}
