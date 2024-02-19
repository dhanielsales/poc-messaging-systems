package main

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nats-io/nats.go"
)

func main() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	nc, err := nats.Connect(os.Getenv("NATS_URL"))
	if err != nil {
		fmt.Println(err)
	}

	defer func() {
		fmt.Println("\nClosing producer...")
		nc.Close()
		close(sigChan)
	}()

	go func() {
		for {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("Enter message: ")
			text, _ := reader.ReadString('\n')
			if text == "exit\n" {
				break
			}

			sanitizedPayload := strings.ReplaceAll(text, "\n", "")
			nc.Publish(os.Getenv("NATS_SUBJECT"), []byte(sanitizedPayload))
		}
	}()

	<-sigChan
}
