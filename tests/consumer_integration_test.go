package tests

import (
	"database/sql"
	"encoding/json"
	"log"
	"os"
	"testing"
	"time"

	"github.com/streadway/amqp"
	"github.com/tuatal/altenar_test/consumer/handler"
	"github.com/tuatal/altenar_test/db"
	"github.com/tuatal/altenar_test/internal/models"
	"github.com/tuatal/altenar_test/internal/repository"
)

type testingConsumerWriter struct {
	t *testing.T
}

func (w testingConsumerWriter) Write(p []byte) (n int, err error) {
	w.t.Log(string(p))
	return len(p), nil
}

func setupConsumerTestDB(t *testing.T) *sql.DB {
	cfg := db.NewConfig(
		os.Getenv("TEST_DB_USER"),
		os.Getenv("TEST_DB_PASSWORD"),
		os.Getenv("TEST_DB_HOST"),
		os.Getenv("TEST_DB_PORT"),
		os.Getenv("TEST_DB_NAME"),
		"disable",
	)
	dbConn, err := db.Connect(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
	}
	_, err = dbConn.Exec("TRUNCATE TABLE transactions")
	if err != nil {
		t.Fatalf("Failed to truncate transactions table: %v", err)
	}

	t.Logf("[setupConsumerTestDB] Test DB setup complete")

	return dbConn
}

func TestConsumerIntegration(t *testing.T) {
	log.SetOutput(testingConsumerWriter{t})
	dbConn := setupConsumerTestDB(t)
	defer dbConn.Close()

	rabbitmqURL := os.Getenv("TEST_RABBITMQ_URL")
	queueName := os.Getenv("TEST_RABBITMQ_QUEUE")

	conn, err := amqp.Dial(rabbitmqURL)
	if err != nil {
		t.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to declare queue: %v", err)
	}

	tx := models.Transaction{
		UserID:          42,
		TransactionType: "bet",
		Amount:          150.0,
		Timestamp:       time.Now(),
	}
	body, err := json.Marshal(tx)
	if err != nil {
		t.Fatalf("Failed to marshal transaction: %v", err)
	}

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		t.Fatalf("Failed to publish message: %v", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		t.Fatalf("Failed to register consumer: %v", err)
	}
	repo := repository.NewTransactionRepo(dbConn)

	done := make(chan bool)
	adaptedMsgs := make(chan handler.Delivery)

	go func() {
		for d := range msgs {
			adaptedMsgs <- handler.AmqpDeliveryAdapter{D: d}
			break
		}
		close(adaptedMsgs)
	}()

	go func() {
		err := handler.HandleMessages(adaptedMsgs, func(tx models.Transaction) error {
			return handler.SaveTransaction(dbConn, tx)
		})
		if err != nil {
			t.Errorf("HandleMessages error: %v", err)
		}
		done <- true
	}()

	select {
	case <-done:
		t.Log("[TestConsumerIntegration] Message processing completed")
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout waiting for message processing")
	}

	transactions, err := repo.GetTransactions(int64(tx.UserID), tx.TransactionType)
	if err != nil {
		t.Fatalf("Failed to get transactions: %v", err)
	}
	if len(transactions) == 0 {
		t.Fatal("No transactions found after processing message")
	}

	found := false
	for _, tr := range transactions {
		if tr.Amount == tx.Amount && tr.UserID == tx.UserID {
			found = true
			break
		}
	}
	if !found {
		t.Error("Inserted transaction not found in database")
	}
}
