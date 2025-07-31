package main

import (
	"database/sql"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
	"github.com/tuatal/altenar_test/consumer/handler"
	"github.com/tuatal/altenar_test/db"
)

var (
	rabbitmqURL string
	queueName   string
)

func init() {
	rabbitmqURL = os.Getenv("RABBITMQ_URL")
	queueName = os.Getenv("RABBITMQ_QUEUE")
}

func main() {
	cfg := db.NewConfig(
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
		"disable",
	)

	dbConn, err := db.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer dbConn.Close()

	for {
		err := runConsumer(dbConn)
		if err != nil {
			log.Printf("Consumer error: %v. Reconnecting in 5 seconds...", err)
			time.Sleep(5 * time.Second)
		}
	}
}

func declareQueue(ch *amqp.Channel, name string) (amqp.Queue, error) {
	return ch.QueueDeclare(
		name,
		true,
		false,
		false,
		false,
		nil,
	)
}

func consumeMessages(ch *amqp.Channel, queueName string) (<-chan amqp.Delivery, error) {
	return ch.Consume(
		queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
}

func runConsumer(db *sql.DB) error {
	log.Println("[runConsumer] Connecting to RabbitMQ at", rabbitmqURL)
	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		log.Printf("[runConsumer] Failed to connect to RabbitMQ: %v", err)
		return err
	}
	defer func() {
		log.Println("[runConsumer] Closing RabbitMQ connection")
		conn.Close()
	}()

	log.Println("[runConsumer] Opening channel")
	ch, err := conn.Channel()
	if err != nil {
		log.Printf("[runConsumer] Failed to open channel: %v", err)
		return err
	}
	defer func() {
		log.Println("[runConsumer] Closing channel")
		ch.Close()
	}()

	log.Println("[runConsumer] Declaring queue:", queueName)
	q, err := declareQueue(ch, queueName)
	if err != nil {
		log.Printf("[runConsumer] Failed to declare queue: %v", err)
		return err
	}

	log.Println("[runConsumer] Consuming messages from queue:", q.Name)
	msgs, err := consumeMessages(ch, q.Name)
	if err != nil {
		log.Printf("[runConsumer] Failed to consume messages: %v", err)
		return err
	}

	log.Println("[runConsumer] Starting message handler")
	err = handler.HandleMessagesProd(msgs, db)
	if err != nil {
		log.Printf("[runConsumer] Handler returned error: %v", err)
	}
	return err
}
