package resolver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/twopow/srd/internal/util"

	cache "github.com/twopow/srd/internal/cache"
)

const (
	// VERSION is the version of the SRD record format
	VERSION = "srd1"
)

var defaultNoHostBaseRedirect = "https://srd.sh"

type ResolverContextKey string

type ResolverConfig struct {
	// RecordingPrefix is the DNS record to use, e.g. "_srd"
	RecordPrefix string

	// InHost is the hostname that should be used for the CNAME record, e.g. "in.srd.sh"
	InHost string

	// InspectorHost is the hostname to be used for the inspector route, e.g. "inspector.srd.sh"
	// if this is empty, inspector will be disabled
	InspectorHost string

	// NoHostBaseRedirect is the URL to redirect to when
	// resolving a request and we fail to find a record
	NoHostBaseRedirect string

	// TTL is the cache TTL
	TTL time.Duration

	// CleanupInterval is how often to cleanup the cache
	CleanupInterval time.Duration

	// Logger is the logger to use
	Logger *slog.Logger
}

type ResolverProvider interface {
	Resolve(ctx context.Context, hostname string) (RR, error)
	Config() *ResolverConfig
	Logger() *slog.Logger
}

type Resolver struct {
	logger *slog.Logger
	cache  cache.CacheProvider
	cfg    ResolverConfig
}

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
var ErrLoop = errors.New("loop detected")
var ErrHostIsIp = errors.New("host is ip")

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

func New(cfg ResolverConfig) (ResolverProvider, error) {
	if cfg.Logger == nil {
		return nil, fmt.Errorf("slog logger is required")
	}

	if cfg.NoHostBaseRedirect == "" {
		cfg.NoHostBaseRedirect = defaultNoHostBaseRedirect
	}

	c, err := cache.New(cache.CacheConfig{
		TTL:             cfg.TTL,
		CleanupInterval: cfg.CleanupInterval,
		Logger:          cfg.Logger,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to init resolver: %w", err)
	}

	return &Resolver{
		cfg:    cfg,
		cache:  c,
		logger: cfg.Logger,
	}, nil
}

func (r *Resolver) Resolve(ctx context.Context, hostname string) (record RR, err error) {
	ctx = context.WithValue(ctx, ResolverContextKey("hostname"), hostname)

	stime := time.Now()

	hostname = strings.ToLower(hostname)
	hostname = strings.TrimSpace(hostname)

	l := r.logger.With("hostname", hostname)

	if util.IsIp(hostname) {
		l.Info("hostname is ip")
		return RR{}, ErrHostIsIp
	}

	if cached, ok := r.getCached(l, hostname); ok {
		l.Info("resolved host",
			"to", cached.To,
			"cached", true,
			"elapsed", time.Since(stime).Milliseconds(),
			"preserveRoute", cached.PreserveRoute,
			"code", cached.Code,
			"referrerPolicy", cached.RefererPolicy.String(),
		)

		return cached, nil
	}

	record, err = r.doResolve(ctx, l, hostname)
	if err != nil {
		return record, err
	}

	l = l.With(
		"to", record.To,
		"elapsed", time.Since(stime).Milliseconds(),
		"preserveRoute", record.PreserveRoute,
		"refererPolicy", record.RefererPolicy.String(),
		"code", record.Code,
	)

	err = r.detectLoop(l, hostname, record.To)
	if err != nil {
		if errors.Is(err, ErrLoop) {
			l.Warn("loop detected")
			return record, ErrLoop
		}

		l.Error("loop detection failed", "error", err)
		return record, fmt.Errorf("loop detection failed: %w", err)
	}

	l.Info("resolved host")
	r.cache.Set(hostname, record)

	return record, nil
}

func (r *Resolver) doResolve(ctx context.Context, l *slog.Logger, hostname string) (record RR, err error) {
	record.NotFound = true
	txtRecords, err := r.resolveTXT(ctx, hostname)

	if err != nil {
		l.Error("failed to resolve host", "error", err)
		return record, err
	}

	if len(txtRecords) == 0 {
		l.Info("no records found")
		return record, nil
	}

	record, err = parseRecord(txtRecords[0])
	if err != nil {
		l.Error("failed to parse record", "error", err)
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
	} else {
		rr.To = strings.ToLower(rr.To)
		rr.To = strings.TrimSpace(rr.To)
	}

	if _, err := url.Parse(rr.To); err != nil {
		return RRNotFound, fmt.Errorf("invalid destination")
	}

	return rr, nil
}

// detectLoop checks if the to host is already in the cache
// if it is, it returns true, otherwise it returns false
func (r *Resolver) detectLoop(l *slog.Logger, hostname, to string) error {
	url, err := url.Parse(to)
	if err != nil {
		return err
	}

	toHost := url.Host

	if toHost == hostname {
		return ErrLoop
	}

	rr, ok := r.getCached(l, toHost)

	// if the record does not exist or is not valid, we are not in a loop
	if !ok || rr.NotFound {
		return nil
	}

	return ErrLoop
}

// resolveTXT takes a hostname with prefix and returns its TXT records.
// Returns an error if the lookup fails
func (r *Resolver) resolveTXT(ctx context.Context, hostname string) ([]string, error) {
	hostname = fmt.Sprintf("%s.%s", r.cfg.RecordPrefix, hostname)

	records, err := net.DefaultResolver.LookupTXT(ctx, hostname)

	if err != nil {
		if dnsErr, ok := err.(*net.DNSError); ok && dnsErr.IsNotFound {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to lookup TXT records for %s: %w", hostname, err)
	}

	return records, nil
}

func (r *Resolver) getCached(l *slog.Logger, hostname string) (rr RR, ok bool) {
	cached, ok := r.cache.Get(hostname)

	if !ok {
		return rr, false
	}

	// cast cached to RR
	if val, ok := cached.(RR); !ok {
		l.Error("invalid cached value, expected RR")
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

func (r *Resolver) Logger() *slog.Logger {
	return r.logger
}

func (r *Resolver) Config() *ResolverConfig {
	return &r.cfg
}
