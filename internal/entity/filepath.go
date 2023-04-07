package entity

import (
	"fmt"
	"os"
)

// A FilePath represents path to a file on local filesystem.
type FilePath string

// Set validates that provided path exists and assigns it to FilePath.
// Required by pflags interface.
func (p *FilePath) Set(src string) error {
	if _, err := os.Stat(src); err != nil {
		return fmt.Errorf("invalid file path: %w", err)
	}

	*p = FilePath(src)

	return nil
}

// String returns string representation of stored address.
// Required by pflags interface.
func (p FilePath) String() string {
	return string(p)
}

// Type returns underlying type used to store FilePath value.
// Required by pflags interface.
func (p FilePath) Type() string {
	return "string"
}
