package main

import (
	"fmt"
	"log"
	"moviedb/carga"
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
		carga.GeneralCharge()
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
	// carga.GeneralCharge()
	// log.Println("PROCESS COMPLETE")
	cronCharge()
	listen()
}
