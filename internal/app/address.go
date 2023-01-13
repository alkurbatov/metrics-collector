package app

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/alkurbatov/metrics-collector/internal/entity"
)

type NetAddress string

func setAddressError(reason error) error {
	return fmt.Errorf("set address failed: %w", reason)
}

func (a *NetAddress) Set(src string) error {
	chunks := strings.Split(src, ":")
	if len(chunks) != 2 {
		return setAddressError(entity.ErrBadAddressFormat)
	}

	if _, err := strconv.Atoi(chunks[1]); err != nil {
		return setAddressError(err)
	}

	*a = NetAddress(src)

	return nil
}

func (a NetAddress) String() string {
	return string(a)
}

func (a NetAddress) Type() string {
	return "string"
}
