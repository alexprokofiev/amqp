package amqp

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type connection = amqp.Connection

type Connection struct {
	*connection

	settings *ConnectionSettings
	isClosed atomic.Bool
}

func NewConnection(
	ctx context.Context,
	amqpURI string,
	options ...connectionOption,
) (*Connection, error) {
	conn := &Connection{
		settings: applyConnectionOptions(options...),
	}

	err := conn.dial(amqpURI)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	go conn.handleReconnect(ctx, amqpURI)

	return conn, nil
}

func (conn *Connection) dial(amqpURI string) (err error) {
	conn.connection, err = amqp.DialConfig(
		amqpURI,
		amqp.Config{
			Heartbeat: conn.settings.heartbeat,
		},
	)
	if err == nil {
		conn.isClosed.Store(false)
	}

	return
}

func (conn *Connection) IsClosed() bool {
	return conn.isClosed.Load()
}

func (conn *Connection) channel() (*amqp.Channel, error) {
	if conn.isClosed.Load() {
		return nil, amqp.ErrClosed
	}

	return conn.connection.Channel()
}

func (conn *Connection) Channel(ctx context.Context, options ...channelOption) (*Channel, error) {
	return NewChannel(ctx, conn, options...)
}

func (conn *Connection) handleReconnect(
	ctx context.Context,
	amqpURI string,
) {
	var (
		notifyConnClose = conn.connection.NotifyClose(make(chan *amqp.Error, 1))
		err             error
	)

	for {
		select {
		case <-ctx.Done():
			return

		case amqpErr, ok := <-notifyConnClose:
			if !ok {
				continue
			}

			conn.isClosed.Store(true)

			for range conn.settings.maxReconnectRetries {
				if err = conn.dial(amqpURI); err == nil {
					break
				}

				<-time.After(conn.settings.reconnectInterval)
			}

			if err != nil {
				amqp.Logger.Printf("failed to reconnect to rmq: %s: %s", amqpErr.Error(), err.Error())

				return
			}

			notifyConnClose = conn.connection.NotifyClose(make(chan *amqp.Error, 1))
		}
	}
}
