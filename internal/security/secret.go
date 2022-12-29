package security

import "strings"

type Secret string

func (s Secret) String() string {
	return strings.Repeat("*", len(s))
}
