package models

import "time"

type Transaction struct {
	UserID          int       `json:"user_id"`
	TransactionType string    `json:"transaction_type"`
	Amount          float64   `json:"amount"`
	Timestamp       time.Time `json:"timestamp"`
}
