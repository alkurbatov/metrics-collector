package mypkg

import "os"

func shutdown() {
	code := 0

	if code == 0 {
		os.Exit(0)
	}

	os.Exit(code)
}
