package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/tuatal/altenar_test/api"
	"github.com/tuatal/altenar_test/db"
	"github.com/tuatal/altenar_test/internal/repository"
	"github.com/tuatal/altenar_test/internal/service"
)

func waitForDB(cfg *db.Config, maxRetries int, delay time.Duration) (*sql.DB, error) {
	var dbConn *sql.DB
	var err error
	for i := 0; i < maxRetries; i++ {
		dbConn, err = db.Connect(cfg)
		if err == nil {
			return dbConn, nil
		}
		log.Printf("Waiting for DB connection... attempt %d/%d: %v", i+1, maxRetries, err)
		time.Sleep(delay)
	}
	return nil, err
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

	dbConn, err := waitForDB(cfg, 10, 3*time.Second)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer dbConn.Close()

	repo := repository.NewTransactionRepo(dbConn)
	svc := service.NewTransactionService(repo)
	handler := &api.Handler{Service: svc}

	r := mux.NewRouter()
	r.HandleFunc("/transactions", handler.GetTransactions).Methods("GET")

	log.Println("API server started on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
