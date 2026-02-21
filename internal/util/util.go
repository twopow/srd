package util

import (
	"fmt"
	"log/slog"
	"regexp"

	"github.com/google/uuid"
)

var IpRegex = regexp.MustCompile(`^(?:[0-9]{1,3}\.){3}[0-9]{1,3}(?::[0-9]{1,5})?$`)

func UUID7() uuid.UUID {
	u, err := uuid.NewV7()
	if err != nil {
		slog.Error("failed to generate uuid", "error", err)
		panic(fmt.Errorf("failed to generate uuid: %w", err))
	}

	return u
}

func IsIp(hostname string) bool {
	return IpRegex.MatchString(hostname)
}
