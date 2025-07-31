package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/tuatal/altenar_test/api"
	"github.com/tuatal/altenar_test/internal/models"
)

type mockService struct {
	transactions []models.Transaction
}

func (m *mockService) GetTransactions(userID int64, txType string) ([]models.Transaction, error) {
	var filtered []models.Transaction
	for _, tx := range m.transactions {
		if (userID == 0 || int64(tx.UserID) == userID) && (txType == "all" || tx.TransactionType == txType) {
			filtered = append(filtered, tx)
		}
	}
	return filtered, nil
}

func TestGetTransactions(t *testing.T) {
	handler := &api.Handler{
		Service: &mockService{
			transactions: []models.Transaction{
				{UserID: 1, TransactionType: "bet", Amount: 100.0},
				{UserID: 2, TransactionType: "win", Amount: 50.0},
			},
		},
	}

	req := httptest.NewRequest("GET", "/transactions?user_id=1&transaction_type=bet", nil)
	w := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/transactions", handler.GetTransactions).Methods("GET")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var transactions []models.Transaction
	if err := json.NewDecoder(w.Body).Decode(&transactions); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(transactions))
	}

	tx := transactions[0]
	if tx.UserID != 1 || tx.TransactionType != "bet" || tx.Amount != 100.0 {
		t.Errorf("Unexpected transaction data: %+v", tx)
	}
}

func TestGetTransactions_NonExistentUser(t *testing.T) {
	handler := &api.Handler{
		Service: &mockService{
			transactions: []models.Transaction{
				{UserID: 1, TransactionType: "bet", Amount: 100.0},
			},
		},
	}

	req := httptest.NewRequest("GET", "/transactions?user_id=9999", nil)
	w := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/transactions", handler.GetTransactions).Methods("GET")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", w.Code)
	}

	var transactions []models.Transaction
	if err := json.NewDecoder(w.Body).Decode(&transactions); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(transactions) != 0 {
		t.Fatalf("Expected 0 transactions, got %d", len(transactions))
	}
}

func TestGetTransactions_InvalidTypeParam(t *testing.T) {
	handler := &api.Handler{Service: &mockService{}}

	req := httptest.NewRequest("GET", "/transactions?transaction_type=invalid", nil)
	w := httptest.NewRecorder()

	r := mux.NewRouter()
	r.HandleFunc("/transactions", handler.GetTransactions).Methods("GET")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d", w.Code)
	}
}
