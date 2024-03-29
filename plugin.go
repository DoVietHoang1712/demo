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
	"time"

	"github.com/IncSW/geoip2"
	"github.com/robfig/cron"
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
	update         string
}

// New creates a new plugin handler.
func New(_ context.Context, next http.Handler, cfg *Config, name string) (http.Handler, error) {
	go func() {
		UpdateDatabase()
	}()
	if len(cfg.DisallowedASNs) > 0 && len(cfg.AllowedASNs) > 0 {
		return nil, errors.New("either allowed asn or disallowed asn could be set at once")
	}
	config := LoadConfig()
	log.Println(config.License)
	if !cfg.Enabled {
		log.Printf("%s: disabled", name)

		return &Plugin{
			next: next,
			name: name,
		}, nil
	}

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
	log.Println(cfg)
	lookup = CreateASNDBLookup(rdr)
	return &Plugin{
		next:           next,
		name:           name,
		enabled:        cfg.Enabled,
		allowedASNs:    cfg.AllowedASNs,
		disallowedASNs: cfg.DisallowedASNs,
		header:         cfg.Header,
		lookup:         lookup,
	}, nil
}

// ServeHTTP implements http.Handler interface.
func (p *Plugin) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if !p.enabled {
		p.next.ServeHTTP(rw, req)
		return
	}
	ips := p.GetRemoteIPs(req, p.header)
	for _, ip := range ips {
		ipDetail, err := p.CheckAllowed(ip)
		if err != nil {
			if errors.Is(err, ErrNotAllowed) {
				log.Printf("%s: %s - access denied for %s (%v)", p.name, req.Host, ip, ipDetail)
			} else {
				log.Printf("%s: %s - %v", p.name, req.Host, err)
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

func UpdateDatabase() {
	config := LoadConfig()
	c := cron.New()
	c.AddFunc(config.CronExpression, func() {
		fileZip := DownloadZipFile("file")
		r, err := os.Open(fileZip)
		if err != nil {
			log.Printf("Open: Open file failed: %s", err.Error())
		}
		folderName := unzipSource(r)
		os.Rename(folderName+"GeoLite2-ASN.mmdb", "GeoLite2-ASN.mmdb")
		name := folderName[0 : len(folderName)-1]
		time.Sleep(1 * time.Second)
		os.RemoveAll(name)
		os.RemoveAll(fileZip)
	})
}
