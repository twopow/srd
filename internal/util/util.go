package util

import (
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

func UUID7() uuid.UUID {
	u, err := uuid.NewV7()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to generate uuid")
	}

	return u
}
