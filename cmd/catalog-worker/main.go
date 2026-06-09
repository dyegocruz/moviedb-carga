package main

import (
	"log"

	"moviedb/internal/app"
)

var runCatalogWorker = app.RunCatalogWorker
var mainErrorHandler = func(err error) {
	log.Fatal(err)
}

func run() error {
	return runCatalogWorker()
}

func main() {
	if err := run(); err != nil {
		mainErrorHandler(err)
	}
}
