package main

import (
	"context"
	"flag"
	"log"

	"github.com/alexprokofiev/amqp"
	amqp091 "github.com/rabbitmq/amqp091-go"
)

var (
	uri   = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	queue = flag.String("queue", "test-queue", "Ephemeral AMQP queue name")
)

func init() {
	flag.Parse()
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	amqpConnection, err := amqp.NewConnection(ctx, *uri)
	if err != nil {
		log.Fatalf("failed to connect to rmq: %s", err.Error())
	}

	defer amqpConnection.Close()

	channel, err := amqpConnection.Channel(ctx)
	if err != nil {
		log.Fatalf("failed to create channel: %s", err.Error())
	}

	err = channel.PublishWithContext(
		ctx,
		"",
		*queue,
		false,
		false,
		amqp091.Publishing{
			ContentType: "text/plain",
			Body:        []byte("hello world"),
		},
	)
	if err != nil {
		log.Fatalf("failed to publish message: %s", err.Error())
	}
}
