package app

import (
	"errors"
	"strconv"
	"strings"
)

type NetAddress string

func (a *NetAddress) Set(src string) error {
	chunks := strings.Split(src, ":")
	if len(chunks) != 2 {
		return errors.New("expected address in a host:port form, got " + src)
	}

	if _, err := strconv.Atoi(chunks[1]); err != nil {
		return err
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
