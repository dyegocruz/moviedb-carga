package main

import (
	"log"

	"moviedb/internal/app"
)

func main() {
	if err := app.RunCatalogWorker(); err != nil {
		log.Fatal(err)
	}
}
