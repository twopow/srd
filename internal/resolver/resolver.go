package resolver

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"srd/internal/log"
	"srd/internal/util"

	cacheM "srd/internal/cache"
)

const (
	VERSION = "srd1"
)

type ResolverConfig struct {
	RecordPrefix       string `help:"Record prefix." default:"_srd"`
	NoHostBaseRedirect string `help:"No host base redirect." default:"https://github.com/twopow/srd"`
}

var (
	DefaultResolver ResolverProvider
)

type ResolverProvider interface {
	Resolve(hostname string) (RR, error)
}

type Resolver struct {
	cache cacheM.CacheProvider
	cfg   ResolverConfig
}

var ErrLoop = errors.New("loop detected")

type RefererPolicy int

const (
	// return no referer header
	RefererPolicyNone RefererPolicy = iota

	// return a referer header with the hostname
	RefererPolicyHost

	// return a referer header with the full url
	RefererPolicyFull
)

func (r RefererPolicy) String() string {
	return []string{"none", "host", "full"}[r]
}

var DefaultRefererPolicy = RefererPolicyHost

// RR is a Redirect Record
type RR struct {
	Hostname      string
	To            string
	PreserveRoute bool
	RefererPolicy RefererPolicy
	Code          int
	NotFound      bool
	Version       string
}

var RRNotFound = RR{NotFound: true, RefererPolicy: RefererPolicyNone, Code: http.StatusNotFound}

func Init(_cfg ResolverConfig, _cache cacheM.CacheProvider) {
	DefaultResolver = &Resolver{
		cache: _cache,
		cfg:   _cfg,
	}
}

func (r *Resolver) Resolve(hostname string) (record RR, err error) {
	stime := time.Now()
	l := log.With("hostname", hostname)

	// if hostname is ip, return the default redirect
	if r.cfg.NoHostBaseRedirect != "" && util.IsIp(hostname) {
		l.Info().Msg("no host base redirect")
		return RR{To: r.cfg.NoHostBaseRedirect}, nil
	}

	if cached, ok := r.getCached(l, hostname); ok {
		l.Info().WithMap(map[string]any{
			"to":             cached.To,
			"cached":         true,
			"elapsed":        time.Since(stime).Milliseconds(),
			"preserveRoute":  cached.PreserveRoute,
			"code":           cached.Code,
			"referrerPolicy": cached.RefererPolicy.String(),
		}).Msg("resolved host")

		return cached, nil
	}

	record, err = r.doResolve(l, hostname)
	if err != nil {
		return record, err
	}

	l = l.WithMap(map[string]any{
		"to":            record.To,
		"elapsed":       time.Since(stime).Milliseconds(),
		"preserveRoute": record.PreserveRoute,
		"refererPolicy": record.RefererPolicy.String(),
		"code":          record.Code,
	})

	err = r.detectLoop(l, record.To)
	if err != nil {
		if errors.Is(err, ErrLoop) {
			l.Warn().Msg("loop detected")
			return record, ErrLoop
		}

		l.Error().Err(err).Msg("loop detection failed")
		return record, fmt.Errorf("loop detection failed: %w", err)
	}

	l.Info().Msg("resolved host")
	r.cache.Set(hostname, record)

	return record, nil
}

func (r *Resolver) doResolve(l *log.Logger, hostname string) (record RR, err error) {
	record.NotFound = true
	txtRecords, err := r.resolveTXT(hostname)

	if err != nil {
		l.Error().Err(err).Msg("failed to resolve host")
		return record, err
	}

	if len(txtRecords) == 0 {
		l.Info().Msg("no records found")
		return record, nil
	}

	record, err = parseRecord(txtRecords[0])
	if err != nil {
		l.Error().Err(err).Msg("failed to parse record")
		return record, err
	}

	// url.Parse expects a scheme
	if !strings.Contains(record.To, "://") {
		record.To = "http://" + record.To
	}

	record.Hostname = hostname
	return record, nil
}

func parseRecord(record string) (RR, error) {
	rr := RR{
		NotFound:      false,
		Code:          http.StatusFound,
		RefererPolicy: DefaultRefererPolicy,
	}

	// remove bounding quotes if they exist
	record = strings.Trim(record, "\"")

	parts := strings.Split(record, ";")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		segments := strings.Split(part, "=")
		if len(segments) == 0 {
			continue // Skip malformed parts
		}

		key := strings.TrimSpace(segments[0])
		value := ""

		if len(segments) > 1 {
			value = strings.TrimSpace(segments[1])
		}

		switch key {
		case "v":
			rr.Version = value
		case "dest":
			rr.To = value
		case "code":
			rr.Code = parseCode(value)
		case "route":
			if value == "preserve" {
				rr.PreserveRoute = true
			}
		case "referer":
			rr.RefererPolicy = parseRefererPolicy(value)
		case "referrer":
			rr.RefererPolicy = parseRefererPolicy(value)
		}
	}

	if rr.Version != VERSION {
		return RRNotFound, fmt.Errorf("invalid version")
	}

	if rr.To == "" {
		return RRNotFound, fmt.Errorf("no destination found")
	}

	if _, err := url.Parse(rr.To); err != nil {
		return RRNotFound, fmt.Errorf("invalid destination")
	}

	return rr, nil
}

// detectLoop checks if the to host is already in the cache
// if it is, it returns true, otherwise it returns false
func (r *Resolver) detectLoop(l *log.Logger, to string) error {
	url, err := url.Parse(to)
	if err != nil {
		return err
	}

	toHost := url.Host

	_, ok := r.getCached(l, toHost)
	if !ok {
		return nil
	}

	return ErrLoop
}

// resolveTXT takes a hostname with prefix and returns its TXT records.
// Returns an error if the lookup fails
func (r *Resolver) resolveTXT(hostname string) ([]string, error) {
	hostname = fmt.Sprintf("%s.%s", r.cfg.RecordPrefix, hostname)

	records, err := net.LookupTXT(hostname)

	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to lookup TXT records for %s: %w", hostname, err)
	}

	return records, nil
}

func (r *Resolver) getCached(l *log.Logger, hostname string) (rr RR, ok bool) {
	cached, ok := r.cache.Get(hostname)

	if !ok {
		return rr, false
	}

	// cast cached to RR
	if val, ok := cached.(RR); !ok {
		l.Error().Msg("invalid cached value, expected RR")
		return rr, false
	} else {
		return val, true
	}
}

// parseCode parses the code string and returns the corresponding http status code
func parseCode(code string) int {
	switch code {
	case "301":
		return http.StatusMovedPermanently
	case "302":
		return http.StatusFound
	case "307":
		return http.StatusTemporaryRedirect
	case "308":
		return http.StatusPermanentRedirect
	default:
		return http.StatusFound
	}
}

func parseRefererPolicy(policy string) RefererPolicy {
	switch policy {
	case "none":
		return RefererPolicyNone
	case "host":
		return RefererPolicyHost
	case "full":
		return RefererPolicyFull
	default:
		return DefaultRefererPolicy
	}
}
