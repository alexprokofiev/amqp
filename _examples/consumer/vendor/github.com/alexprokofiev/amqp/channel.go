package amqp

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type channel = amqp.Channel

type Channel struct {
	*channel

	closed   bool
	settings *ChannelSettings
}

func NewChannel(ctx context.Context, conn *Connection, options ...channelOption) (*Channel, error) {
	ch, err := conn.channel()
	if err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	channel := &Channel{
		channel:  ch,
		settings: applyChannelOptions(options...),
	}

	if err = channel.applySettings(); err != nil {
		return nil, fmt.Errorf("failed to apply settings: %w", err)
	}

	go channel.handleReconnect(ctx, conn)

	return channel, nil
}

func (channel *Channel) applySettings() (err error) {
	if err = channel.channel.Qos(
		channel.settings.QosPrefetchCount,
		channel.settings.QosPrefetchSize,
		channel.settings.QosGlobal,
	); err != nil {
		err = fmt.Errorf("failed to set qos: %w", err)
	}

	return
}

func (channel *Channel) IsClosed() bool {
	return channel == nil || channel.closed
}

func (channel *Channel) Close() error {
	if channel.IsClosed() {
		return amqp.ErrClosed
	}

	channel.closed = true

	return channel.channel.Close()
}

func (channel *Channel) handleReconnect(ctx context.Context, conn *Connection) {
	defer func() {
		if channel.IsClosed() {
			return
		}

		err := channel.Close()
		if err != nil {
			amqp.Logger.Printf("failed to close channel: %s", err.Error())
		}
	}()

	var (
		notifyClose = channel.NotifyClose(make(chan *amqp.Error, 1))
		err         error
	)

	for {
		select {
		case <-ctx.Done():
			return

		case amqpErr := <-notifyClose:
			channel.closed = true

			for range conn.settings.maxReconnectRetries {
				if !conn.IsClosed() {
					channel.channel, err = conn.channel()
					if err == nil {
						channel.closed = false

						err = channel.channel.Qos(
							channel.settings.QosPrefetchCount,
							channel.settings.QosPrefetchSize,
							channel.settings.QosGlobal,
						)
						if err != nil {
							amqp.Logger.Printf("failed to set qos: %s", err.Error())
						}

						break
					}
				}

				<-time.After(conn.settings.reconnectInterval)
			}

			if err != nil {
				amqp.Logger.Printf("fatal: failed to recreate channel: %s: %s", amqpErr.Error(), err.Error())

				return
			}

			notifyClose = channel.NotifyClose(make(chan *amqp.Error, 1))
		}
	}
}
