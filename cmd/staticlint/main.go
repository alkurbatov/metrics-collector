package main

import "github.com/alkurbatov/metrics-collector/pkg/staticlint"

func main() {
	app := staticlint.New()
	app.Run()
}
