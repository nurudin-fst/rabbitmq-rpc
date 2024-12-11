package main

import (
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type User struct {
	UserID   string
	Name     string
	Username string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func handleRequest(ch *amqp.Channel, msg amqp.Delivery) {
	// Proses request
	user := &User{
		UserID:   msg.CorrelationId,
		Name:     "Kenan",
		Username: "kenanfst",
	}
	fmt.Printf("decode %+v\n", user)

	response, error := json.Marshal(user)
	if error != nil {
		panic(error)
	}

	// Kirim respons kembali ke client
	err := ch.Publish(
		"",          // Exchange
		msg.ReplyTo, // Reply-to queue
		false,       // Mandatory
		false,       // Immediate
		amqp.Publishing{
			CorrelationId: msg.CorrelationId,
			ContentType:   "application/json",
			Body:          []byte(response),
		},
	)
	failOnError(err, "Failed to publish a message")
}

func main() {
	// Koneksi ke RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Mendeklarasikan antrean untuk menerima request
	queue, err := ch.QueueDeclare(
		"rpc_queue", // Nama antrean
		false,       // Durable
		false,       // Auto-delete
		false,       // Exclusive
		false,       // No-wait
		nil,         // Arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Mendengarkan antrean
	msgs, err := ch.Consume(
		queue.Name, // Nama queue
		"",         // Consumer tag
		true,       // Auto-ack
		false,      // Exclusive
		false,      // No-local
		false,      // No-wait
		nil,        // Arguments
	)
	failOnError(err, "Failed to register a consumer")

	// Proses request yang diterima
	for msg := range msgs {
		go handleRequest(ch, msg) // Proses permintaan di goroutine untuk menangani banyak request
	}
}
