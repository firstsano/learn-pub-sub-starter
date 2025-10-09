package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	rabbitConn := "amqp://guest:guest@localhost:5672/"
	rabbit, err := amqp.Dial(rabbitConn)
	if err != nil {
		log.Fatal(err)
	}
	defer rabbit.Close()

	fmt.Println("Connected to RabbitMQ")
	fmt.Println("Starting Peril server...")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan

	fmt.Println("Received an interrupt, stopping services...")
}
