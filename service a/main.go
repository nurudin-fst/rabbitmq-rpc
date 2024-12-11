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

type ResponseOrder struct {
	OrderId int32
	Vehicle string
	User
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	// Koneksi ke RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// Menyiapkan reply queue
	replyQueue, err := ch.QueueDeclare(
		"",    // Nama queue kosong agar RabbitMQ otomatis memberi nama
		false, // Durable
		true,  // Auto-delete
		false, // Exclusive
		false, // No-wait
		nil,   // Arguments
	)
	failOnError(err, "Failed to declare a queue")

	// Membuat channel untuk menerima respons
	msgs, err := ch.Consume(
		replyQueue.Name, // Nama queue
		"",              // Consumer tag
		true,            // Auto-ack
		false,           // Exclusive
		false,           // No-local
		false,           // No-wait
		nil,             // Arguments
	)
	fmt.Println(replyQueue.Name)
	failOnError(err, "Failed to register a consumer")

	// Mengirimkan request ke service B
	corrID := "12345"                  // ID unik untuk request
	message := "Hello from Service A!" // Pesan request
	err = ch.Publish(
		"",          // Exchange
		"rpc_queue", // Routing key
		false,       // Mandatory
		false,       // Immediate
		amqp.Publishing{
			CorrelationId: corrID,
			ReplyTo:       replyQueue.Name,
			ContentType:   "text/plain",
			Body:          []byte(message),
		},
	)
	failOnError(err, "Failed to publish a message")

	// Menerima respons dari layanan B
	for msg := range msgs {
		var user User
		if err := json.Unmarshal(msg.Body, &user); err != nil {
			panic(err)
		}
		response := ResponseOrder{
			OrderId: 1,
			Vehicle: "motor",
			User: User{
				UserID:   user.UserID,
				Name:     user.Name,
				Username: user.Username,
			},
		}

		if msg.CorrelationId == corrID {
			fmt.Printf("response %+v", response)
			break
		}
	}
}
