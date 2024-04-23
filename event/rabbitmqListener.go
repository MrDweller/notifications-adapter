package event

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

const EXCHANGE = "exchange"

type RabbitmqListener struct {
	address  string
	port     int
	event    EventDefinition
	metadata map[string]string

	done chan bool
}

func NewRabbitmqListener(address string, port int, event EventDefinition, metadata map[string]string) *RabbitmqListener {
	done := make(chan bool)
	return &RabbitmqListener{
		address:  address,
		port:     port,
		event:    event,
		metadata: metadata,

		done: done,
	}
}

func (listener *RabbitmqListener) GetListenerId() string {
	return fmt.Sprintf("%s-%s", listener.event, listener.metadata[EXCHANGE])
}

func (listener *RabbitmqListener) Listen(output chan<- []byte) error {
	url := fmt.Sprintf("%s:%d/", listener.address, listener.port)
	dialAddrr := fmt.Sprintf("amqp://guest:guest@%s", url)
	conn, err := amqp.Dial(dialAddrr)
	if err != nil {
		return err
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	err = ch.ExchangeDeclare(
		listener.metadata[EXCHANGE], // name
		"fanout",                    // type
		true,                        // durable
		false,                       // auto-deleted
		false,                       // internal
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		return err
	}

	q, err := ch.QueueDeclare(
		"",    // name
		false, // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return err
	}

	err = ch.QueueBind(
		q.Name,                      // queue name
		"",                          // routing key
		listener.metadata[EXCHANGE], // exchange
		false,                       // no-wait
		nil,                         // arguments
	)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			output <- d.Body
		}
	}()

	<-listener.done
	return nil

}

func (listener *RabbitmqListener) Stop() error {
	listener.done <- true
	return nil
}
