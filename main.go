package main

import (
	"fmt"
	"log"

	catalogCharge "moviedb/catalog-charge"
	"moviedb/configs"
	"moviedb/database"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/robfig/cron"
)

func init() {
	database.CheckCreateCollections()
}

func cronCharge() {
	c := cron.New()
	c.AddFunc("@daily", func() {
		log.Println("[Job] General Charge")
		catalogCharge.GeneralCharge()
		log.Println("PROCESS COMPLETE")
	})
	log.Println("Start Job")
	c.Start()
}

func listen() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	fmt.Println(time.Now().String() + " - Closed")
}

func main() {
	if configs.IsProduction() {
		cronCharge()
		listen()
	} else {
		catalogCharge.GeneralCharge()
		log.Println("PROCESS COMPLETE")
	}
}
