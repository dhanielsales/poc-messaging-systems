package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nats-io/nats.go"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	///////

	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		fmt.Println(err)
	}

	defer func() {
		fmt.Println("\nClosing consumer...")
		nc.Close()
		close(sigChan)
	}()

	go func() {
		subject := os.Getenv("NATS_SUBJECT")
		queue := os.Getenv("NATS_QUEUE")

		nc.QueueSubscribe(subject, queue, func(m *nats.Msg) {
			fmt.Println(m.Sub.Queue)
			fmt.Printf("Received a message: %s\n", string(m.Data))
		})
	}()

	///////
	<-sigChan
}
