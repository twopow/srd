package resolver

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"strings"
	"time"

	"srd/internal/log"

	cacheM "srd/internal/cache"
	"srd/internal/config"
)

const (
	VERSION = "srd1"
)

var (
	DefaultResolver ResolverProvider
)

type ResolverProvider interface {
	Resolve(hostname string) (RR, error)
}

type Resolver struct {
	cache cacheM.CacheProvider
	cfg   config.ResolverConfig
}

var ipRegex = regexp.MustCompile(`^[0-9\.:]+$`)

// RR is a Redirect Record
type RR struct {
	Hostname string
	To       string
	NotFound bool
	Version  string
}

func Init(_cfg config.ResolverConfig, _cache cacheM.CacheProvider) {
	DefaultResolver = &Resolver{
		cache: _cache,
		cfg:   _cfg,
	}
}

func (r *Resolver) Resolve(hostname string) (record RR, err error) {
	stime := time.Now()
	l := log.With("hostname", hostname)

	// if hostname is ip, return the default redirect
	if r.cfg.NoHostBaseRedirect != "" && ipRegex.MatchString(hostname) {
		l.Info().Msg("no host base redirect")
		return RR{To: r.cfg.NoHostBaseRedirect}, nil
	}

	if cached, ok := r.getCached(l, hostname); ok {
		l.Info().
			Str("to", cached.To).
			Int64("elapsed", time.Since(stime).Milliseconds()).
			Msg("resolved host")

		return cached, nil
	}

	record, err = r.doResolve(l, hostname)
	if err != nil {
		return record, err
	}

	l.Info().
		Str("to", record.To).
		Int64("elapsed", time.Since(stime).Milliseconds()).
		Msg("resolved host")

	r.cache.Set(hostname, record)

	return record, nil
}

func (r *Resolver) doResolve(l *log.Logger, hostname string) (record RR, err error) {
	l.Info().Msg("resolving hostname")

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

	record.Hostname = hostname
	return record, nil
}

func parseRecord(record string) (RR, error) {
	rr := RR{
		NotFound: false,
	}

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
		}
	}

	if rr.Version != VERSION {
		return RR{NotFound: true}, fmt.Errorf("invalid version")
	}

	if rr.To == "" {
		return RR{NotFound: true}, fmt.Errorf("no destination found")
	}

	if _, err := url.Parse(rr.To); err != nil {
		return RR{NotFound: true}, fmt.Errorf("invalid destination")
	}

	return rr, nil
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

	l.Info().Msg("cache hit")

	rr = cached.(RR)
	return rr, true
}
