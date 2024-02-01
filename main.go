package main

import (
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

		for _, message := range output.Messages {
			chn <- message
		}

	}

}

func main() {

	catalogCharge.CatalogSearchCharge()
	// CATALOG SEARCH TV
	// catalogTv := tv.GetCatalogSearch()
	// log.Println(len(catalogTv))
	// catalogTvLocalizated := make(map[int]catalogCharge.CatalogSearch, 0)
	// for _, item := range catalogTv {
	// 	var catalog catalogCharge.CatalogSearch
	// 	if catalogTvLocalizated[item.Id].Id == 0 {
	// 		catalog.Id = item.Id
	// 		catalog.CatalogType = common.MEDIA_TYPE_TV
	// 		catalog.FirstAirDate = item.FirstAirDate
	// 		catalog.OriginalLanguage = item.OriginalLanguage
	// 		catalog.OriginalTitle = item.OriginalTitle
	// 		catalog.Popularity = item.Popularity
	// 		catalogTvLocalizated[item.Id] = catalog
	// 	}

	// 	var location catalogCharge.Location
	// 	location.Language = item.Language
	// 	location.Title = item.Title
	// 	location.PosterPath = item.PosterPath

	// 	loc := catalogTvLocalizated[item.Id]
	// 	loc.Locations = append(loc.Locations, location)
	// 	catalogTvLocalizated[item.Id] = loc
	// }

	// for _, item := range catalogTvLocalizated {
	// 	log.Println(item)
	// 	// req := elastic.NewBulkIndexRequest().
	// 	// 	Index(newIndexName).
	// 	// 	Doc(item)
	// 	// bulkProcessor.Add(req)
	// }

	log.Println("PROCESS COMPLETE")

	// if configs.IsProduction() {
	// 	cronCharge()
	// } else {
	// 	catalogCharge.GeneralCharge()
	// 	log.Println("PROCESS COMPLETE")
	// }

	// chnMessages := make(chan *sqs.Message, 1)
	// go pollMessages(chnMessages)

	// for message := range chnMessages {
	// 	var esChargeMessage queue.EsChargeMessage
	// 	json.Unmarshal([]byte(*message.Body), &esChargeMessage)

	// 	if esChargeMessage.Env == configs.GetEnv() {
	// 		receiptHandle := message.ReceiptHandle
	// 		err := queue.DeleteMessage(configs.GetQueueUrl(), receiptHandle)
	// 		if err != nil {
	// 			fmt.Printf("Got an error while trying to delete message: %v", err)
	// 			return
	// 		}

	// 		catalogCharge.ElasticGeneralCharge()
	// 	} else {
	// 		log.Println("No messages for this environment")
	// 	}
	// }

}
