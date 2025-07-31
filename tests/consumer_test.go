package tests

import (
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/tuatal/altenar_test/consumer/handler"
	"github.com/tuatal/altenar_test/internal/models"
)

type mockDelivery struct {
	body       []byte
	ackCalled  bool
	nackCalled bool
	mu         sync.Mutex
}

func (m *mockDelivery) Ack(multiple bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ackCalled = true
	return nil
}

func (m *mockDelivery) Nack(multiple, requeue bool) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nackCalled = true
	return nil
}

func (m *mockDelivery) Body() []byte {
	return m.body
}

func TestHandleMessages(t *testing.T) {
	var saved []models.Transaction
	var mu sync.Mutex

	saveFunc := func(tx models.Transaction) error {
		mu.Lock()
		defer mu.Unlock()
		saved = append(saved, tx)
		return nil
	}

	msgs := make(chan handler.Delivery)

	go func() {
		err := handler.HandleMessages(msgs, saveFunc)
		if err != nil {
			t.Errorf("HandleMessages error: %v", err)
		}
	}()

	tx := models.Transaction{
		UserID:          1,
		TransactionType: "bet",
		Amount:          100.0,
		Timestamp:       time.Now(),
	}
	body, _ := json.Marshal(tx)

	validMsg := &mockDelivery{body: body}
	invalidMsg := &mockDelivery{body: []byte("invalid json")}

	msgs <- validMsg
	msgs <- invalidMsg
	close(msgs)

	time.Sleep(100 * time.Millisecond)

	if !validMsg.ackCalled {
		t.Error("Expected Ack to be called for valid message")
	}
	if !invalidMsg.nackCalled {
		t.Error("Expected Nack to be called for invalid message")
	}

	mu.Lock()
	if len(saved) != 1 {
		t.Errorf("Expected 1 saved transaction, got %d", len(saved))
	}
	mu.Unlock()
}

func TestHandleMessages_SaveError(t *testing.T) {
	saveFunc := func(tx models.Transaction) error {
		return errors.New("save error")
	}
	msgs := make(chan handler.Delivery)

	go func() {
		err := handler.HandleMessages(msgs, saveFunc)
		if err != nil {
			t.Errorf("HandleMessages error: %v", err)
		}
	}()

	tx := models.Transaction{
		UserID:          1,
		TransactionType: "bet",
		Amount:          100.0,
		Timestamp:       time.Now(),
	}
	body, _ := json.Marshal(tx)
	validMsg := &mockDelivery{body: body}

	msgs <- validMsg
	close(msgs)

	time.Sleep(100 * time.Millisecond)
}

func TestHandleMessages_MultipleMessages(t *testing.T) {
	var saved []models.Transaction
	var mu sync.Mutex

	saveFunc := func(tx models.Transaction) error {
		mu.Lock()
		defer mu.Unlock()
		saved = append(saved, tx)
		return nil
	}

	msgs := make(chan handler.Delivery)

	go func() {
		err := handler.HandleMessages(msgs, saveFunc)
		if err != nil {
			t.Errorf("HandleMessages error: %v", err)
		}
	}()

	for i := 0; i < 5; i++ {
		tx := models.Transaction{
			UserID:          int(i),
			TransactionType: "bet",
			Amount:          float64(i * 10),
			Timestamp:       time.Now(),
		}
		body, _ := json.Marshal(tx)
		msgs <- &mockDelivery{body: body}
	}
	close(msgs)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if len(saved) != 5 {
		t.Errorf("Expected 5 saved transactions, got %d", len(saved))
	}
	mu.Unlock()
}

func TestHandleMessages_InvalidJSON(t *testing.T) {
	var saved []models.Transaction
	var mu sync.Mutex

	saveFunc := func(tx models.Transaction) error {
		mu.Lock()
		defer mu.Unlock()
		saved = append(saved, tx)
		return nil
	}

	msgs := make(chan handler.Delivery)

	go func() {
		err := handler.HandleMessages(msgs, saveFunc)
		if err != nil {
			t.Errorf("HandleMessages error: %v", err)
		}
	}()

	msgs <- &mockDelivery{body: []byte("invalid json")}
	close(msgs)

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	if len(saved) != 0 {
		t.Errorf("Expected 0 saved transactions, got %d", len(saved))
	}
	mu.Unlock()
}
