package util

import (
	"regexp"
	"srd/internal/log"

	"github.com/google/uuid"
)

var IpRegex = regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}(?::[0-9]{1,5})?$`)

func UUID7() uuid.UUID {
	u, err := uuid.NewV7()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to generate uuid")
	}

	return u
}

func IsIp(hostname string) bool {
	return IpRegex.MatchString(hostname)
}
