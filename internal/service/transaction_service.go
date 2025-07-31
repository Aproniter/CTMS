package service

import (
	"errors"

	"github.com/tuatal/altenar_test/internal/models"
)

type TransactionServiceInterface interface {
	GetTransactions(userID int64, txType string) ([]models.Transaction, error)
}

type TransactionService struct {
	repo TransactionServiceInterface
}

func NewTransactionService(repo TransactionServiceInterface) *TransactionService {
	return &TransactionService{repo: repo}
}

func (s *TransactionService) GetTransactions(userID int64, txType string) ([]models.Transaction, error) {
	if txType != "bet" && txType != "win" && txType != "all" {
		return nil, errors.New("invalid transaction type")
	}
	return s.repo.GetTransactions(userID, txType)
}
