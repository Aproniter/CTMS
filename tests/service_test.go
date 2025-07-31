package tests

import (
	"errors"
	"testing"

	"github.com/tuatal/altenar_test/internal/models"
	"github.com/tuatal/altenar_test/internal/service"
)

type mockRepo struct {
	transactions []models.Transaction
	err          error
}

func (m *mockRepo) GetTransactions(userID int64, txType string) ([]models.Transaction, error) {
	return m.transactions, m.err
}

func TestGetTransactions_ValidType(t *testing.T) {
	mock := &mockRepo{
		transactions: []models.Transaction{
			{UserID: 1, TransactionType: "bet", Amount: 100},
		},
		err: nil,
	}
	svc := service.NewTransactionService(mock)

	txs, err := svc.GetTransactions(1, "bet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(txs) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(txs))
	}
}

func TestGetTransactions_InvalidType(t *testing.T) {
	mock := &mockRepo{
		transactions: []models.Transaction{
			{UserID: 1, TransactionType: "bet", Amount: 100},
		},
		err: nil,
	}
	svc := service.NewTransactionService(mock)

	_, err := svc.GetTransactions(1, "invalid")
	if err == nil {
		t.Fatal("expected error for invalid transaction type")
	}
	if err.Error() != "invalid transaction type" {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestGetTransactions_RepoError(t *testing.T) {
	mock := &mockRepo{
		transactions: nil,
		err:          errors.New("db error"),
	}
	svc := service.NewTransactionService(mock)

	_, err := svc.GetTransactions(1, "bet")
	if err == nil {
		t.Fatal("expected error from repo")
	}
	if err.Error() != "db error" {
		t.Fatalf("unexpected error message: %v", err)
	}
}
