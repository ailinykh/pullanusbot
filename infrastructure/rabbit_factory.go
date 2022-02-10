package infrastructure

import (
	"github.com/ailinykh/pullanusbot/v2/core"
	"github.com/streadway/amqp"
)

func CreateRabbitFactory(l core.ILogger, url string) (core.ITaskFactory, func()) {
	conn, err := amqp.Dial(url)
	if err != nil {
		panic(err)
	}

	go func() {
		err := <-conn.NotifyClose(make(chan *amqp.Error))
		l.Error("connection closed", err)
	}()

	ch, err := conn.Channel()
	if err != nil {
		panic(err)
	}

	return &RabbitFactory{l, conn, ch}, func() {
		ch.Close()
		conn.Close()
	}
}

type RabbitFactory struct {
	l    core.ILogger
	conn *amqp.Connection
	ch   *amqp.Channel
}

func (q *RabbitFactory) NewTask(name string) core.ITask {
	return &RabbitWorker{q.l, name, q.ch}
}
