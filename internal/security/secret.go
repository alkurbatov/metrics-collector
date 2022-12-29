package security

import (
	"regexp"
	"strings"
)

type Secret string

func (s Secret) String() string {
	return strings.Repeat("*", len(s))
}

var databaseSecrets = regexp.MustCompile(`(://).*:.*(@)`)

type DatabaseURL string

func (u DatabaseURL) String() string {
	return string(databaseSecrets.ReplaceAll([]byte(u), []byte("$1*****:*****$2")))
}
