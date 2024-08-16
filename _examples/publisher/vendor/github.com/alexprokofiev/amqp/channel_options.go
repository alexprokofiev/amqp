package amqp

type (
	ChannelSettings struct {
		QosPrefetchCount int
		QosPrefetchSize  int
		QosGlobal        bool
	}

	channelOption func(*ChannelSettings) *ChannelSettings
)

var defaultChannelSettings = &ChannelSettings{
	QosPrefetchCount: 0,
	QosPrefetchSize:  0,
	QosGlobal:        false,
}

func applyChannelOptions(options ...channelOption) *ChannelSettings {
	settings := defaultChannelSettings

	for _, option := range options {
		settings = option(settings)
	}

	return settings
}

func WithQosPrefetchCount(count int) channelOption {
	return func(settings *ChannelSettings) *ChannelSettings {
		settings.QosPrefetchCount = count

		return settings
	}
}

func WithQosPrefetchSize(size int) channelOption {
	return func(settings *ChannelSettings) *ChannelSettings {
		settings.QosPrefetchSize = size

		return settings
	}
}

func WithQosGlobal(global bool) channelOption {
	return func(settings *ChannelSettings) *ChannelSettings {
		settings.QosGlobal = global

		return settings
	}
}
