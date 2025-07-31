package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/tuatal/altenar_test/internal/service"
)

type Handler struct {
	Service service.TransactionServiceInterface
}

func (h *Handler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	userIDStr := query.Get("user_id")
	var userID int64
	var err error
	if userIDStr != "" {
		userID, err = strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user_id", http.StatusBadRequest)
			return
		}
	}
	txType := query.Get("transaction_type")
	if txType == "" {
		txType = "all"
	}
	if txType != "bet" && txType != "win" && txType != "all" {
		http.Error(w, "Invalid transaction_type", http.StatusBadRequest)
		return
	}
	transactions, err := h.Service.GetTransactions(userID, txType)
	if err != nil {
		http.Error(w, "Failed to get transactions", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transactions)
}
