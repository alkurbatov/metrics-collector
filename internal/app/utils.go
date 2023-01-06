package app

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// Log if panic occurres but try to avoid program termination.
func tryRecover() {
	if p := recover(); p != nil {
		log.Error().Err(fmt.Errorf("%v", p)).Msg("") //nolint: goerr113
		log.Debug().Stack()
	}
}
