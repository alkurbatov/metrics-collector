package main

import "os"

func main() {
	go os.Exit(1) // want "os.Exit in main function is forbidden"
}
