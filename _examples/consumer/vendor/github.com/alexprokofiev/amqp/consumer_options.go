package amqp

import amqp "github.com/rabbitmq/amqp091-go"

type (
	ConsumerSettings struct {
		*ChannelSettings
		*QueueSettings

		ConsumerName string
		AutoAck      bool
		Exclusive    bool
		NoLocal      bool
		Args         amqp.Table
	}

	QueueSettings struct {
		Name       string
		Durable    bool
		AutoDelete bool
		Exclusive  bool
		Args       amqp.Table
	}

	consumerOption func(settings *ConsumerSettings) *ConsumerSettings
)

var (
	defaultConsumerSettings = &ConsumerSettings{
		ConsumerName:    "",
		AutoAck:         false,
		Exclusive:       false,
		NoLocal:         false,
		Args:            nil,
		ChannelSettings: defaultChannelSettings,
		QueueSettings:   defaultQueueSettings,
	}

	defaultQueueSettings = &QueueSettings{
		Name:       "",
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		Args:       nil,
	}
)

func applyConsumerOptions(options ...consumerOption) *ConsumerSettings {
	settings := defaultConsumerSettings

	for _, option := range options {
		settings = option(settings)
	}

	return settings
}

func WithConsumerQueueName(queueName string) consumerOption {
	return func(settings *ConsumerSettings) *ConsumerSettings {
		settings.QueueSettings.Name = queueName

		return settings
	}
}

func WithConsumerQueueDurable(durable bool) consumerOption {
	return func(settings *ConsumerSettings) *ConsumerSettings {
		settings.QueueSettings.Durable = durable

		return settings
	}
}

func WithConsumerQueueAutoDelete(autoDelete bool) consumerOption {
	return func(settings *ConsumerSettings) *ConsumerSettings {
		settings.QueueSettings.AutoDelete = autoDelete

		return settings
	}
}

func WithConsumerQueueExclusive(exclusive bool) consumerOption {
	return func(settings *ConsumerSettings) *ConsumerSettings {
		settings.QueueSettings.Exclusive = exclusive

		return settings
	}
}

func WithConsumerQueueArgs(args amqp.Table) consumerOption {
	return func(settings *ConsumerSettings) *ConsumerSettings {
		settings.QueueSettings.Args = args

		return settings
	}
}

func WithConsumerName(consumerName string) consumerOption {
	return func(settings *ConsumerSettings) *ConsumerSettings {
		settings.ConsumerName = consumerName

		return settings
	}
}

func WithAutoAck(autoAck bool) consumerOption {
	return func(settings *ConsumerSettings) *ConsumerSettings {
		settings.AutoAck = autoAck

		return settings
	}
}

func WithExclusive(exclusive bool) consumerOption {
	return func(settings *ConsumerSettings) *ConsumerSettings {
		settings.Exclusive = exclusive

		return settings
	}
}

func WithNoLocal(noLocal bool) consumerOption {
	return func(settings *ConsumerSettings) *ConsumerSettings {
		settings.NoLocal = noLocal

		return settings
	}
}

func WithArgs(args amqp.Table) consumerOption {
	return func(settings *ConsumerSettings) *ConsumerSettings {
		settings.Args = args

		return settings
	}
}
