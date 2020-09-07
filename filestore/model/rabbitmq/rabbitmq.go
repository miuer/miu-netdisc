package rabbitmq

import (
	"database/sql"
	"log"

	"github.com/miuer/miu-netdisc/filestore/config"
	"github.com/streadway/amqp"
)

// TransferMeta -
type TransferMeta struct {
	FileName     string
	FileSha1     string
	FileSize     int64
	FileCurAddr  string
	FileDestAddr string
}

// RabbitMQ -
type RabbitMQ struct {
	Connection   *amqp.Connection
	Channel      *amqp.Channel
	ExchangeName string
	ExchangeType string
	QueueName    string
	RoutingKey   string
}

// InitRabbitMq -
func initRabbitMq() (r *RabbitMQ) {
	conn, err := amqp.Dial(config.RabbitURL)
	if err != nil {
		log.Fatalln(err)
	}

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalln(err)
	}

	err = channel.ExchangeDeclare(
		config.ExchangeName,
		config.ExchangeType,
		true,
		false,
		false,
		true,
		nil,
	)

	_, err = channel.QueueDeclare(
		config.QueueName,
		true,
		false,
		false,
		true,
		nil,
	)

	err = channel.QueueBind(
		config.QueueName,
		config.RoutingKey,
		config.ExchangeName,
		true,
		nil,
	)

	r = &RabbitMQ{
		conn,
		channel,
		config.ExchangeName,
		config.ExchangeType,
		config.QueueName,
		config.RoutingKey,
	}

	return r

}

// Publish -
func Publish(data []byte) (err error) {
	r := initRabbitMq()

	err = r.Channel.Publish(
		r.ExchangeName,
		r.RoutingKey,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        data,
		},
	)

	r.Channel.Close()
	r.Connection.Close()

	return err
}

// Consume -
func Consume(writer *sql.DB, callback func(*sql.DB, []byte) error) {
	r := initRabbitMq()

	msg, err := r.Channel.Consume(
		r.QueueName,
		"oss",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalln(err)
	}

	done := make(chan bool)

	go func() {
		for v := range msg {
			callback(writer, v.Body)
		}
	}()

	<-done

	r.Channel.Close()
	r.Connection.Close()
}
