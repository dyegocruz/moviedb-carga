package main

import (
	"log"

	"moviedb/internal/app"
)

var runCatalog = app.RunCatalog
var mainErrorHandler = func(err error) {
	log.Fatal(err)
}

func run() error {
	return runCatalog()
}

func main() {
	if err := run(); err != nil {
		mainErrorHandler(err)
	}
}
