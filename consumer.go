package amqp

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	Messages chan *amqp.Delivery

	conn     *Connection
	messages <-chan amqp.Delivery
	channel  *Channel
	settings *ConsumerSettings
}

func NewConsumer(
	ctx context.Context,
	conn *Connection,
	options ...consumerOption,
) (*Consumer, error) {
	var (
		consumer = &Consumer{
			Messages: make(chan *amqp.Delivery, 1),
			conn:     conn,
			settings: applyConsumerOptions(options...),
		}

		err error
	)

	consumer.channel, err = conn.Channel(
		ctx,
		WithQosPrefetchCount(consumer.settings.QosPrefetchCount),
		WithQosPrefetchSize(consumer.settings.QosPrefetchSize),
		WithQosGlobal(consumer.settings.QosGlobal),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	_, err = consumer.channel.QueueDeclare(
		consumer.settings.QueueSettings.Name,
		consumer.settings.QueueSettings.Durable,
		consumer.settings.QueueSettings.AutoDelete,
		consumer.settings.QueueSettings.Exclusive,
		false,
		consumer.settings.QueueSettings.Args,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	consumer.messages, err = consumer.channel.ConsumeWithContext(
		ctx,
		consumer.settings.QueueSettings.Name,
		consumer.settings.ConsumerName,
		consumer.settings.AutoAck,
		consumer.settings.Exclusive,
		consumer.settings.NoLocal,
		false,
		consumer.settings.Args,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume messages: %w", err)
	}

	go consumer.handle(ctx)

	return consumer, nil
}

func (consumer *Consumer) Consume() <-chan *amqp.Delivery {
	return consumer.Messages
}

func (consumer *Consumer) handle(ctx context.Context) {
	var (
		notifyClose = consumer.channel.NotifyClose(make(chan *amqp.Error, 1))
		err         error
	)

	defer close(consumer.Messages)

	for {
		select {
		case <-ctx.Done():
			return

		case message, ok := <-consumer.messages:
			if !ok {
				continue
			}

			consumer.Messages <- &message

		case amqpErr := <-notifyClose:
			for range consumer.conn.settings.maxReconnectRetries {
				if !consumer.channel.IsClosed() {
					consumer.messages, err = consumer.channel.ConsumeWithContext(
						ctx,
						consumer.settings.QueueSettings.Name,
						consumer.settings.ConsumerName,
						consumer.settings.AutoAck,
						consumer.settings.Exclusive,
						consumer.settings.NoLocal,
						false,
						consumer.settings.Args,
					)
					if err == nil {
						break
					}
				}

				<-time.After(consumer.conn.settings.reconnectInterval)
			}

			if err != nil {
				amqp.Logger.Printf("consumer: failed to recreate channel: %s: %s", amqpErr.Error(), err.Error())

				return
			}

			notifyClose = consumer.channel.NotifyClose(make(chan *amqp.Error, 1))
		}
	}
}
