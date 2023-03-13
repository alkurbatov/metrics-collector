package main

import "os"

func main() {
	for i := 0; i < 5; i++ {
		if i%2 == 0 {
			os.Exit(i) // want "os.Exit in main function is forbidden"
		}
	}
}
