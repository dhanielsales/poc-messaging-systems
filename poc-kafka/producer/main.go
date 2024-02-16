package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": "localhost"})
	if err != nil {
		panic(err)
	}

	defer func() {
		fmt.Println("Closing consumer")
		p.Close()
		close(sigChan)
	}()

	topic := os.Getenv("KAFKA_TOPIC")

	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter message: ")
			text, _ := reader.ReadString('\n')
			if text == "exit\n" {
				break
			}

			sanitizedPayload := strings.ReplaceAll(text, "\n", "")
			p.Produce(&kafka.Message{
				TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
				Value:          []byte(sanitizedPayload),
			}, nil)
		}

		p.Flush(15 * 1000)
	}()

	<-sigChan
}
