package main

import (
	"log"

	"moviedb/internal/app"
)

func main() {
	if err := app.RunCatalog(); err != nil {
		log.Fatal(err)
	}
}
