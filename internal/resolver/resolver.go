package resolver

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	cacheM "srd/internal/cache"
	"srd/internal/config"
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

// RR is a Redirect Record
type RR struct {
	Hostname string
	To       string
	NotFound bool
}

func Init(_cfg config.ResolverConfig, _cache cacheM.CacheProvider) {
	DefaultResolver = &Resolver{
		cache: _cache,
		cfg:   _cfg,
	}
}

func (r *Resolver) Resolve(hostname string) (record RR, err error) {
	stime := time.Now()
	l := log.With().
		Str("hostname", hostname).
		Logger()

	if cached, ok := r.getCached(&l, hostname); ok {
		l.Info().
			Str("to", cached.To).
			Int64("elapsed", time.Since(stime).Milliseconds()).
			Msg("resolved host")

		return cached, nil
	}

	record, err = r.doResolve(&l, hostname)

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

func (r *Resolver) doResolve(l *zerolog.Logger, hostname string) (record RR, err error) {
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

	record, err = r.parseRecord(hostname, txtRecords[0])
	if err != nil {
		l.Error().Err(err).Msg("failed to parse record")
		return record, err
	}

	return record, nil
}

func (r *Resolver) parseRecord(hostname string, record string) (RR, error) {
	parts := strings.Split(record, "=")

	if len(parts) != 2 || parts[0] != "to" {
		return RR{}, fmt.Errorf("invalid record format")
	}

	return RR{
		Hostname: hostname,
		To:       parts[1],
		NotFound: false,
	}, nil
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

func (r *Resolver) getCached(l *zerolog.Logger, hostname string) (rr RR, ok bool) {
	cached, ok := r.cache.Get(hostname)

	if !ok {
		return rr, false
	}

	l.Info().Msg("cache hit")

	rr = cached.(RR)
	return rr, true
}
