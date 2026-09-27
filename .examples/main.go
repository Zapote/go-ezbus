package main

import (
	"context"
	"encoding/json"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/zapote/go-ezbus"
	"github.com/zapote/go-ezbus/logger"
	"github.com/zapote/go-ezbus/rabbitmq"
)

type greeting struct {
	Text string `json:"text"`
}

func main() {
	logger.SetLevel(logger.DebugLevel)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	//setup publisher
	bp := rabbitmq.NewBroker("sample-publisher")
	rp := ezbus.NewRouter()
	publisher := ezbus.NewBus(bp, rp)
	err := publisher.Go()
	if err != nil {
		log.Fatalf("Start publisher: %s", err.Error())
	}
	defer publisher.Stop()

	//setup receiver
	br := rabbitmq.NewBroker("sample-receiver")
	rr := ezbus.NewRouter()
	rr.Handle("greeting", handler)
	receiver := ezbus.NewBus(br, rr)
	receiver.SubscribeMessage("sample-publisher", "greeting")

	//publish messsage
	go publish(ctx, publisher)

	//take messages until the process is told to stop
	if err := receiver.Run(ctx); err != nil {
		log.Fatalf("Receiver: %s", err.Error())
	}
}

func publish(ctx context.Context, publisher ezbus.Publisher) {
	for ctx.Err() == nil {
		err := publisher.PublishContext(ctx, greeting{"hello ezbus"})
		if err != nil {
			logger.Error(err.Error())
		} else {
			logger.Info("Message published")
		}
		time.Sleep(time.Second * 3)
	}
}

func handler(m ezbus.Message) error {
	var g greeting
	json.Unmarshal(m.Body, &g)
	logger.Info(g.Text)
	return nil
}
