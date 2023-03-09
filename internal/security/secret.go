package security

import (
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// A Secret is designed to store sensitive data (e.g. passwords)
// and avoid leaking of values during logging.
type Secret string

// Set assigns provided value to Secret.
func (s *Secret) Set(src string) error {
	if len([]byte(src)) < 32 {
		log.Warn().Msg("Insecure signature: secret key is shorter than 32 bytes!")
	}

	*s = Secret(src)

	return nil
}

// Type returns underlying type used to store NetAddress value.
// Required by pflags interface.
func (s Secret) Type() string {
	return "string"
}

// String returns masked representation of stored value.
// Required by pflags interface.
func (s Secret) String() string {
	return strings.Repeat("*", len(s))
}

var databaseSecrets = regexp.MustCompile(`(://).*:.*(@)`)

// A DatabaseURL is designed to store database connection URLs
// and avoid leaking of login and password values during logging.
type DatabaseURL string

// String returns masked representation of stored value.
func (u DatabaseURL) String() string {
	return string(databaseSecrets.ReplaceAll([]byte(u), []byte("$1*****:*****$2")))
}
