package tests

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/tuatal/altenar_test/api"
	"github.com/tuatal/altenar_test/db"
	"github.com/tuatal/altenar_test/internal/models"
	"github.com/tuatal/altenar_test/internal/repository"
	"github.com/tuatal/altenar_test/internal/service"
)

type testingWriter struct {
	t *testing.T
}

func (w testingWriter) Write(p []byte) (n int, err error) {
	w.t.Log(string(p))
	return len(p), nil
}

func setupApiTestDB(t *testing.T) *sql.DB {
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

	return dbConn
}

func insertTestTransaction(t *testing.T, db *sql.DB, tx models.Transaction) {
	query := `
         INSERT INTO transactions (user_id, transaction_type, amount, timestamp)
         VALUES ($1, $2, $3, $4)
     `
	_, err := db.Exec(query, tx.UserID, tx.TransactionType, tx.Amount, tx.Timestamp)
	if err != nil {
		t.Fatalf("Failed to insert test transaction: %v", err)
	}
	t.Logf("Inserted test transaction: %+v", tx)
}

func TestAPIIntegration(t *testing.T) {
	log.SetOutput(testingWriter{t})
	dbConn := setupApiTestDB(t)
	defer dbConn.Close()

	insertTestTransaction(t, dbConn, models.Transaction{
		UserID:          1,
		TransactionType: "bet",
		Amount:          100.0,
		Timestamp:       time.Now(),
	})

	repo := repository.NewTransactionRepo(dbConn)
	svc := service.NewTransactionService(repo)
	handler := &api.Handler{Service: svc}

	r := mux.NewRouter()
	r.HandleFunc("/transactions", handler.GetTransactions).Methods("GET")

	server := httptest.NewServer(r)
	defer server.Close()

	t.Logf("Starting test server at %s", server.URL)

	resp, err := http.Get(fmt.Sprintf("%s/transactions?user_id=1&transaction_type=bet", server.URL))
	if err != nil {
		t.Fatalf("Failed to send GET request: %v", err)
	}
	defer resp.Body.Close()

	t.Logf("Received response with status: %s", resp.Status)

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	var transactions []models.Transaction
	if err := json.NewDecoder(resp.Body).Decode(&transactions); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	t.Logf("Decoded transactions: %+v", transactions)

	if len(transactions) != 1 {
		t.Fatalf("Expected 1 transaction, got %d", len(transactions))
	}

	tx := transactions[0]
	if tx.UserID != 1 || tx.TransactionType != "bet" || tx.Amount != 100.0 {
		t.Errorf("Unexpected transaction data: %+v", tx)
	}
}
