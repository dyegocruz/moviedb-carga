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

	// idsMovies := database.GetAllIdsByLanguage(database.COLLECTION_MOVIE, "en")
	// log.Println(len(idsMovies))

	// var i int64 = 0
	// var interval = int64(1000)
	// var listIdsIn []int = []int{}
	// for i = 0; i < int64(len(idsMovies)); i++ {
	// 	listIdsIn = append(listIdsIn, idsMovies[i])
	// 	if i%interval == 0 {
	// 		log.Println(len(listIdsIn))
	// 		docs := movie.GetCatalogSearchIn(listIdsIn)
	// 		log.Println(i, len(docs))
	// 		listIdsIn = []int{}
	// 	}
	// }

	catalogCharge.CatalogSearchCharge()

	// docsCount := database.GetCountAllByColletcionAndLanguage(database.COLLECTION_PERSON, "en")
	// log.Println(docsCount)

	// var i int64
	// var interval = int64(1000)
	// for i = 0; i < docsCount; i++ {

	// 	if i%interval == 0 {
	// 		docs := person.GetCatalogSearchTest(i, interval)
	// 		log.Println(i, len(docs))
	// 		for _, doc := range docs {
	// 			// req := elastic.NewBulkIndexRequest().
	// 			// 	Index(newIndexName).
	// 			// 	Doc(doc)
	// 			// bulkProcessor.Add(req)
	// 		}
	// 	}
	// }

	// catalogPerson := person.GetCatalogSearch()
	// for _, item := range catalogPerson {
	// 	var catalog catalogCharge.CatalogSearch
	// 	catalog.Id = item.Id
	// 	catalog.Name = item.Name
	// 	catalog.CatalogType = common.MEDIA_TYPE_PERSON
	// 	catalog.ProfilePath = item.ProfilePath
	// 	catalog.Popularity = item.Popularity
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
