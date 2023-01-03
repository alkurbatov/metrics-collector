package security

import (
	"regexp"
	"strings"

	"github.com/alkurbatov/metrics-collector/internal/logging"
)

type Secret string

func (s *Secret) Set(src string) error {
	if len([]byte(src)) < 32 {
		logging.Log.Warning("Insecure signature: secret key is shorter than 32 bytes!")
	}

	*s = Secret(src)

	return nil
}

func (s Secret) Type() string {
	return "string"
}

func (s Secret) String() string {
	return strings.Repeat("*", len(s))
}

var databaseSecrets = regexp.MustCompile(`(://).*:.*(@)`)

type DatabaseURL string

func (u DatabaseURL) String() string {
	return string(databaseSecrets.ReplaceAll([]byte(u), []byte("$1*****:*****$2")))
}
