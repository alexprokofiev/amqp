package amqp

import "time"

type (
	connectionOption func(conn *ConnectionSettings) *ConnectionSettings

	ConnectionSettings struct {
		maxReconnectRetries int
		reconnectInterval   time.Duration
		heartbeat           time.Duration
	}
)

var defaultConnectionSettings = &ConnectionSettings{
	maxReconnectRetries: 20,
	reconnectInterval:   1 * time.Second,
	heartbeat:           1 * time.Second,
}

func applyConnectionOptions(options ...connectionOption) *ConnectionSettings {
	settings := defaultConnectionSettings

	for _, option := range options {
		settings = option(settings)
	}

	return settings
}

func WithMaxReconnectRetries(count int) connectionOption {
	return func(settings *ConnectionSettings) *ConnectionSettings {
		settings.maxReconnectRetries = count

		return settings
	}
}

func WithReconnectInterval(intreval time.Duration) connectionOption {
	return func(settings *ConnectionSettings) *ConnectionSettings {
		settings.reconnectInterval = intreval

		return settings
	}
}

func WithHeartbeat(count time.Duration) connectionOption {
	return func(settings *ConnectionSettings) *ConnectionSettings {
		settings.heartbeat = count

		return settings
	}
}
