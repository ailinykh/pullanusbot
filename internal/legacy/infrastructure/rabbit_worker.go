package infrastructure

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/streadway/amqp"
)

type RabbitWorker struct {
	key string
	ch  *amqp.Channel
}

func (worker *RabbitWorker) Perform(data []byte, ch chan []byte) error {
	q, err := worker.ch.QueueDeclare(
		"",    // name
		true,  // durable
		false, // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	msgs, err := worker.ch.Consume(
		q.Name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to consume channel: %v", err)
	}

	corrId := uuid.NewString()
	err = worker.ch.Publish(
		"",         // exchange
		worker.key, // routing key
		false,      // mandatory
		false,      // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			CorrelationId: corrId,
			Body:          data,
			ReplyTo:       q.Name,
		})
	if err != nil {
		return fmt.Errorf("failed to publish a task: %v", err)
	}

	go func() {
		for d := range msgs {
			if corrId == d.CorrelationId {
				ch <- d.Body
				break
			}
		}
	}()

	return err
}
