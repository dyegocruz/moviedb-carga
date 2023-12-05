package queue

import (
	"fmt"
	"moviedb/configs"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func getSession() *session.Session {
	creds := credentials.NewStaticCredentials(configs.GetAcessKeyId(), configs.GetSecretAccessKey(), "")
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: "default",
		Config: aws.Config{
			Credentials: creds,
			Region:      aws.String("us-east-1"),
		},
	})

	if err != nil {
		fmt.Printf("Got an error while trying to retrieve session: %v", err)
		return nil
	}

	return sess
}

func SendMessage(queueUrl string, messageBody string) error {
	sqsClient := sqs.New(getSession())

	_, err := sqsClient.SendMessage(&sqs.SendMessageInput{
		QueueUrl:    &queueUrl,
		MessageBody: aws.String(messageBody),
	})

	return err
}

func GetMessages(queueUrl string, maxMessages int) (*sqs.ReceiveMessageOutput, error) {
	sqsClient := sqs.New(getSession())

	msgResult, err := sqsClient.ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            &queueUrl,
		MaxNumberOfMessages: aws.Int64(1),
	})

	if err != nil {
		return nil, err
	}

	return msgResult, nil
}

func DeleteMessage(queueUrl string, messageHandle *string) error {
	sqsClient := sqs.New(getSession())

	_, err := sqsClient.DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      &queueUrl,
		ReceiptHandle: messageHandle,
	})

	return err
}
