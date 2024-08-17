package infrastructure

import (
	"github.com/ailinykh/pullanusbot/v2/internal/core"
	legacy "github.com/ailinykh/pullanusbot/v2/internal/legacy/core"
	"github.com/streadway/amqp"
)

func CreateRabbitFactory(l core.Logger, url string) (legacy.ITaskFactory, func()) {
	factory := &RabbitFactory{l: l}
	err := factory.reestablishConnection(url)
	if err != nil {
		panic(err)
	}
	return factory, factory.Close
}

type RabbitFactory struct {
	l    core.Logger
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewTask is a core.ITaskFactory interface implementation
func (q *RabbitFactory) NewTask(name string) legacy.ITask {
	return &RabbitWorker{name, q.ch}
}

func (q *RabbitFactory) Close() {
	q.ch.Close()
	q.conn.Close()
}

func (q *RabbitFactory) reestablishConnection(url string) error {
	conn, err := amqp.Dial(url)
	if err != nil {
		q.l.Error(err)
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		q.l.Error(err)
		return err
	}

	q.conn = conn
	q.ch = ch

	go func() {
		err := <-conn.NotifyClose(make(chan *amqp.Error))
		q.l.Error("connection closed", err)
		errr := q.reestablishConnection(url)
		if errr != nil {
			q.l.Error(errr)
		}
	}()

	return nil
}
