package main

import "os"

func main() {
	code := 0

	if code == 0 {
		os.Exit(0) // want "os.Exit in main function is forbidden"
	}

	os.Exit(code) // want "os.Exit in main function is forbidden"
}
