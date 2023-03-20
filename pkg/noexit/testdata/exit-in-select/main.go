package main

import (
	"os"
	"time"
)

func main() {
	signal := make(chan os.Signal)
	ticker := time.NewTicker(100)

	defer ticker.Stop()

	select {
	case <-signal:
		os.Exit(0) // want "os.Exit in main function is forbidden"

	case <-ticker.C:
		os.Exit(1) // want "os.Exit in main function is forbidden"
	}
}
