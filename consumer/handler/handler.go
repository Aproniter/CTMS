package handler

import (
	"database/sql"
	"encoding/json"
	"log"

	"github.com/streadway/amqp"
	"github.com/tuatal/altenar_test/internal/models"
)

type Delivery interface {
	Ack(multiple bool) error
	Nack(multiple, requeue bool) error
	Body() []byte
}

type AmqpDeliveryAdapter struct {
	D amqp.Delivery
}

func (a AmqpDeliveryAdapter) Ack(multiple bool) error {
	return a.D.Ack(multiple)
}

func (a AmqpDeliveryAdapter) Nack(multiple, requeue bool) error {
	return a.D.Nack(multiple, requeue)
}

func (a AmqpDeliveryAdapter) Body() []byte {
	return a.D.Body
}

func HandleMessages(msgs <-chan Delivery, saveFunc func(models.Transaction) error) error {
	txChan := make(chan models.Transaction, 100)
	done := make(chan bool)

	go func() {
		for tx := range txChan {
			if err := saveFunc(tx); err != nil {
				log.Printf("[saveTransaction] Failed to save transaction: %v", err)
			} else {
				log.Printf("[saveTransaction] Transaction saved: %+v", tx)
			}
		}
		done <- true
	}()

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var tx models.Transaction
			if err := json.Unmarshal(d.Body(), &tx); err != nil {
				log.Printf("[HandleMessages] Invalid message format: %v", err)
				d.Nack(false, false)
				continue
			}
			log.Printf("[HandleMessages] Received transaction: %+v", tx)
			txChan <- tx
			if err := d.Ack(false); err != nil {
				log.Printf("[HandleMessages] Failed to ack message: %v", err)
			} else {
				log.Printf("[HandleMessages] Message acked")
			}
		}
		log.Println("[HandleMessages] Message channel closed")
		close(txChan)
		<-done
		forever <- true
	}()

	<-forever
	return nil
}

func HandleMessagesProd(msgs <-chan amqp.Delivery, db *sql.DB) error {
	adaptedMsgs := make(chan Delivery)

	go func() {
		for d := range msgs {
			adaptedMsgs <- AmqpDeliveryAdapter{D: d}
		}
		close(adaptedMsgs)
	}()

	return HandleMessages(adaptedMsgs, func(tx models.Transaction) error {
		return SaveTransaction(db, tx)
	})
}

func SaveTransaction(db *sql.DB, tx models.Transaction) error {
	query := `
         INSERT INTO transactions (user_id, transaction_type, amount, timestamp)
         VALUES ($1, $2, $3, $4)
     `
	_, err := db.Exec(query, tx.UserID, tx.TransactionType, tx.Amount, tx.Timestamp)
	if err != nil {
		log.Printf("[saveTransaction] DB exec error: %v", err)
		return err
	}
	log.Printf("[saveTransaction] Inserted transaction: %+v", tx)
	return nil

}
