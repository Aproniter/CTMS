package repository

import (
	"database/sql"
	"strconv"
	"strings"

	"github.com/tuatal/altenar_test/internal/models"
)

type TransactionRepo struct {
	db *sql.DB
}

func NewTransactionRepo(db *sql.DB) *TransactionRepo {
	return &TransactionRepo{db: db}
}

func (r *TransactionRepo) GetTransactions(userID int64, txType string) ([]models.Transaction, error) {
	var args []interface{}
	var conditions []string
	query := "SELECT user_id, transaction_type, amount, timestamp FROM transactions"
	argPos := 1

	if userID != 0 {
		conditions = append(conditions, "user_id = $"+strconv.Itoa(argPos))
		args = append(args, userID)
		argPos++
	}

	if txType != "all" {
		conditions = append(conditions, "transaction_type = $"+strconv.Itoa(argPos))
		args = append(args, txType)
		argPos++
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY timestamp DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []models.Transaction
	for rows.Next() {
		var tx models.Transaction
		if err := rows.Scan(&tx.UserID, &tx.TransactionType, &tx.Amount, &tx.Timestamp); err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	return transactions, nil
}
