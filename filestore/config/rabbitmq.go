package config

const (
	RabbitURL = "amqp://admin:123456@192.168.2.3:5672/"

	ExchangeName = "oss.transfer.exchange"
	ExchangeType = "direct"
	QueueName    = "oss.transfer.queue"
	RoutingKey   = "oss"
)
