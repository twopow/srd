package util

import (
	"srd/internal/log"

	"github.com/google/uuid"
)

func UUID7() uuid.UUID {
	u, err := uuid.NewV7()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to generate uuid")
	}

	return u
}
