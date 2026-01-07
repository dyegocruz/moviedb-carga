package main

import (
	"log"

	catalogCharge "moviedb/catalog-charge"
	"moviedb/database"

	"github.com/robfig/cron"
)

func init() {
	database.CheckCreateCollections()
}

func cronCharge() {
	c := cron.New()
	c.AddFunc("@daily", func() {
		log.Println("[Job] General Catalog Handler")
		catalogCharge.GeneralCatalogHandler()
		log.Println("PROCESS COMPLETE")
	})

	c.AddFunc("0 0 3 * * *", func() {
		log.Println("[Job] Calling: ElasticGeneralCharge Catalog process")
		catalogCharge.ElasticGeneralCharge()
		log.Println("PROCESS COMPLETE")
	})

	log.Println("Start Job")
	c.Start()
}

func main() {
	cronCharge()
	select {}
}
