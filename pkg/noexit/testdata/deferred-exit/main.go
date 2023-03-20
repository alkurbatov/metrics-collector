package main

import "os"

func main() {
	defer os.Exit(1) // want "os.Exit in main function is forbidden"

	defer func() {
		os.Exit(0)
	}()
}
