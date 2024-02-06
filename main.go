package main

import (
	"encoding/json"
	"fmt"
	"log"

	catalogCharge "moviedb/catalog-charge"
	"moviedb/configs"
	"moviedb/database"
	"moviedb/queue"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/service/sqs"
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

func pollMessages(chn chan<- *sqs.Message) {

	for {
		maxMessages := 1
		output, err := queue.GetMessages(configs.GetQueueUrl(), maxMessages)

		if err != nil {
			fmt.Printf("failed to fetch sqs message %v", err)
		}

		// Reduce the quanitity of requests to check the messages
		time.Sleep(10 * time.Minute)

		for _, message := range output.Messages {
			chn <- message
		}

	}

}

func main() {

	if configs.IsProduction() {
		cronCharge()
	} else {
		catalogCharge.GeneralCharge()
		log.Println("PROCESS COMPLETE")
	}

	chnMessages := make(chan *sqs.Message, 1)
	go pollMessages(chnMessages)

	for message := range chnMessages {
		var esChargeMessage queue.EsChargeMessage
		json.Unmarshal([]byte(*message.Body), &esChargeMessage)

		if esChargeMessage.Env == configs.GetEnv() {
			receiptHandle := message.ReceiptHandle
			err := queue.DeleteMessage(configs.GetQueueUrl(), receiptHandle)
			if err != nil {
				fmt.Printf("Got an error while trying to delete message: %v", err)
				return
			}

			catalogCharge.ElasticGeneralCharge()
		} else {
			log.Println("No messages for this environment")
		}
	}

}
