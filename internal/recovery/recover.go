// Package recovery provides panic recovering utility
// which allows to resume goroutine execution if possible.
package recovery

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// TryRecover tries to avoid program termination if panic occures.
func TryRecover() {
	if p := recover(); p != nil {
		log.Error().Err(fmt.Errorf("%v", p)).Msg("") //nolint: goerr113
		log.Debug().Stack()
	}
}
