package tests

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/tuatal/altenar_test/internal/repository"
)

func TestGetTransactionsRepo(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock database: %v", err)
	}
	defer db.Close()

	repo := repository.NewTransactionRepo(db)

	columns := []string{"user_id", "transaction_type", "amount", "timestamp"}

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT user_id, transaction_type, amount, timestamp FROM transactions WHERE user_id = $1 AND transaction_type = $2 ORDER BY timestamp DESC")).
		WithArgs(int64(1), "bet").
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(int64(1), "bet", 100.0, time.Now()))

	txs, err := repo.GetTransactions(1, "bet")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT user_id, transaction_type, amount, timestamp FROM transactions WHERE user_id = $1 ORDER BY timestamp DESC")).
		WithArgs(int64(2)).
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(int64(2), "win", 50.0, time.Now()))

	txs, err = repo.GetTransactions(2, "all")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(txs) != 1 {
		t.Errorf("expected 1 transaction, got %d", len(txs))
	}

	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT user_id, transaction_type, amount, timestamp FROM transactions ORDER BY timestamp DESC")).
		WithArgs().
		WillReturnRows(sqlmock.NewRows(columns).
			AddRow(int64(3), "bet", 200.0, time.Now()).
			AddRow(int64(4), "win", 150.0, time.Now()))

	txs, err = repo.GetTransactions(0, "all")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(txs) != 2 {
		t.Errorf("expected 2 transactions, got %d", len(txs))
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %v", err)
	}
}
