package main

import "os"

func shutdown() {
	os.Exit(0)
}

func main() {
	defer shutdown()
}
