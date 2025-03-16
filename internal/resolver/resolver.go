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

// RR is a Redirect Record
type RR struct {
	Hostname string
	To       string
	NotFound bool
}

var (
	cfg   config.ResolverConfig
	cache *cacheM.Cache
)

func Init(_cfg config.ResolverConfig, _cache *cacheM.Cache) {
	cfg = _cfg
	cache = _cache
}

func Resolve(hostname string) (record RR, err error) {
	stime := time.Now()
	l := log.With().
		Str("hostname", hostname).
		Logger()

	if cached, ok := getCached(&l, hostname); ok {
		l.Info().
			Str("to", cached.To).
			Int64("elapsed", time.Since(stime).Milliseconds()).
			Msg("resolved host")

		return cached, nil
	}

	record, err = doResolve(&l, hostname)

	if err != nil {
		return record, err
	}

	l.Info().
		Str("to", record.To).
		Int64("elapsed", time.Since(stime).Milliseconds()).
		Msg("resolved host")

	cache.Set(hostname, record)

	return record, nil
}

func doResolve(l *zerolog.Logger, hostname string) (record RR, err error) {
	l.Info().Msg("resolving hostname")

	record.NotFound = true
	txtRecords, err := resolveTXT(hostname)

	if err != nil {
		l.Error().Err(err).Msg("failed to resolve host")
		return record, err
	}

	if len(txtRecords) == 0 {
		l.Info().Msg("no records found")
		return record, nil
	}

	record, err = parseRecord(hostname, txtRecords[0])
	if err != nil {
		l.Error().Err(err).Msg("failed to parse record")
		return record, err
	}

	return record, nil
}

func parseRecord(hostname string, record string) (RR, error) {
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
func resolveTXT(hostname string) ([]string, error) {
	hostname = fmt.Sprintf("%s.%s", cfg.RecordPrefix, hostname)

	records, err := net.LookupTXT(hostname)

	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to lookup TXT records for %s: %w", hostname, err)
	}

	return records, nil
}

func getCached(l *zerolog.Logger, hostname string) (rr RR, ok bool) {
	cached, ok := cache.Get(hostname)

	if !ok {
		return rr, false
	}

	l.Info().Msg("cache hit")

	rr = cached.(RR)
	return rr, true
}
