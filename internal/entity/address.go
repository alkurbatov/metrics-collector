package entity

import (
	"fmt"
	"strconv"
	"strings"
)

// A NetAddress represents network address in format ip:port, e.g. 0.0.0.0:8080.
type NetAddress string

func setAddressError(reason error) error {
	return fmt.Errorf("set address failed: %w", reason)
}

// Set validates format of provided value and assigns it to NetAddress.
// Required by pflags interface.
func (a *NetAddress) Set(src string) error {
	chunks := strings.Split(src, ":")
	if len(chunks) != 2 {
		return setAddressError(ErrBadAddressFormat)
	}

	if _, err := strconv.Atoi(chunks[1]); err != nil {
		return setAddressError(err)
	}

	*a = NetAddress(src)

	return nil
}

// String returns string representation of stored address.
// Required by pflags interface.
func (a NetAddress) String() string {
	return string(a)
}

// Type returns underlying type used to store NetAddress value.
// Required by pflags interface.
func (a NetAddress) Type() string {
	return "string"
}
