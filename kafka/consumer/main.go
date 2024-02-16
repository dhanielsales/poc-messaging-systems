package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func setupTopics() {
	partitions := os.Getenv("KAFKA_TOPIC_PARTITIONS")
	topic := os.Getenv("KAFKA_TOPIC")

	c, err := kafka.NewAdminClient(&kafka.ConfigMap{"bootstrap.servers": os.Getenv("KAFKA_HOST")})
	if err != nil {
		panic(err)
	}

	_, err = c.DeleteTopics(context.Background(), []string{topic})
	if err != nil {
		panic(err)
	}

	if partitions == "" {
		partitions = "2"
	}

	p, err := strconv.Atoi(partitions)
	if err != nil {
		panic(err)
	}

	_, err = c.CreateTopics(context.Background(), []kafka.TopicSpecification{{
		Topic:             topic,
		NumPartitions:     p,
		ReplicationFactor: 1,
	}})

	if err != nil {
		panic(err)
	}

	c.Close()
}

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	setupTopics()

	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": os.Getenv("KAFKA_HOST"),
		"group.id":          os.Getenv("KAFKA_GROUP_ID"),
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		panic(err)
	}

	defer func() {
		fmt.Println("Closing consumer")
		c.Close()
		close(sigChan)
	}()

	topic := os.Getenv("KAFKA_TOPIC")

	go func() {
		c.SubscribeTopics([]string{topic}, nil)

		run := true

		for run {
			msg, err := c.ReadMessage(time.Second)
			if err == nil {
				fmt.Printf("Message on %s: %s\n", msg.TopicPartition, string(msg.Value))
			} else if !err.(kafka.Error).IsTimeout() {
				fmt.Printf("Consumer error: %v (%v)\n", err, msg)
			}
		}
	}()

	<-sigChan
}
