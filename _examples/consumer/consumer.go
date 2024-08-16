package main

import (
	"context"
	"flag"
	"log"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/alexprokofiev/amqp"
	"github.com/rabbitmq/amqp091-go"
)

var (
	uri   = flag.String("uri", "amqp://guest:guest@localhost:5672/", "AMQP URI")
	queue = flag.String("queue", "test-queue", "Ephemeral AMQP queue name")
)

func init() {
	flag.Parse()

	amqp091.SetLogger(log.New(os.Stderr, "", 0))
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	amqpConnection, err := amqp.NewConnection(ctx, *uri, amqp.WithHeartbeat(1*time.Second))
	if err != nil {
		log.Fatalf("failed to connect to rmq: %s", err.Error())
	}

	consumer, err := amqp.NewConsumer(
		ctx,
		amqpConnection,
		amqp.WithConsumerQueueName(*queue),
	)
	if err != nil {
		log.Fatalf("failed to create consumer: %s", err.Error())
	}

	for {
		select {
		case <-ctx.Done():
			return

		case message, ok := <-consumer.Consume():
			if !ok {
				log.Fatalf("failed to consume message")
			}

			log.Printf("got message: %s", string(message.Body))

			err = message.Ack(false)
			if err != nil {
				log.Fatalf("failed to ack message: %s", err.Error())
			}

			return
		}
	}
}
