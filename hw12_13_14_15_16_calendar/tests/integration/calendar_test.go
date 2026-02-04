// tests/integration/calendar_test.go
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Event struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	DateTime     string `json:"datetime"`
	Duration     int64  `json:"duration"`
	UserID       string `json:"user_id"`
	NotifyBefore int64  `json:"notify_before,omitempty"`
}

func TestCalendarIntegration(t *testing.T) {
	apiURL := os.Getenv("CALENDAR_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8888"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	now := time.Now().Add(30 * time.Second)
	eventID := "test-integration-1"
	event := Event{
		ID:           eventID,
		Title:        "Integration Test",
		DateTime:     now.Format(time.RFC3339),
		Duration:     3600,
		UserID:       "user1",
		NotifyBefore: 15,
	}

	// === 1. Создание события ===
	body, _ := json.Marshal(event)
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL+"/events", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	defer func() { _ = resp.Body.Close() }()

	// === 2. Проверка бизнес-ошибки: дубликат времени ===
	duplicateEvent := Event{
		ID:           "test-dup",
		Title:        "Duplicate",
		DateTime:     now.Format(time.RFC3339), // то же время
		Duration:     1800,
		UserID:       "user1", // тот же пользователь
		NotifyBefore: 10,
	}
	body, _ = json.Marshal(duplicateEvent)
	req, _ = http.NewRequestWithContext(ctx, "POST", apiURL+"/events", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ = http.DefaultClient.Do(req)
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
	defer func() { _ = resp.Body.Close() }()

	// === 3. Листинги: day / week / month ===
	dateStr := now.Format("2006-01-02")

	// Day
	checkListing(t, apiURL+"/events/day?date="+dateStr, eventID)

	// Week
	checkListing(t, apiURL+"/events/week?start="+dateStr, eventID)

	// Month
	checkListing(t, apiURL+"/events/month?start="+dateStr, eventID)

	// === 4. Проверка отправки уведомления через RabbitMQ ===
	rmqURL := os.Getenv("TEST_RMQ_URL")
	if rmqURL == "" {
		rmqURL = "amqp://calendar:calendar@localhost:5672/"
	}

	var conn *amqp.Connection
	var ch *amqp.Channel

	// Retry loop
	for i := 0; i < 10; i++ {
		conn, err = amqp.Dial(rmqURL) // ← используем =, а не :=
		if err == nil {
			break
		}
		t.Logf("Attempt %d: failed to connect to RabbitMQ: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}
	require.NoError(t, err, "failed to connect to RabbitMQ after retries")
	defer func() { _ = conn.Close() }()

	ch, err = conn.Channel() // ← снова =
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	ch, err = conn.Channel()
	require.NoError(t, err)
	defer func() { _ = ch.Close() }()

	// Объявляем очередь (идемпотентно)
	q, err := ch.QueueDeclare("notifications", true, false, false, false, nil)
	require.NoError(t, err)

	// Получаем одно сообщение
	msgs, err := ch.Consume(q.Name, "", false, false, false, false, nil)
	require.NoError(t, err)

	// Ждём до 15 секунд
	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var received bool
	for {
		select {
		case msg, ok := <-msgs:
			if !ok {
				t.Fatal("RabbitMQ channel closed unexpectedly")
			}
			var notification struct {
				EventID string `json:"event_id"`
			}
			if err := json.Unmarshal(msg.Body, &notification); err != nil {
				t.Logf("failed to parse notification: %v", err)
				continue // или break
			}
			if notification.EventID == eventID {
				received = true
				if err := msg.Ack(false); err != nil { // подтверждаем обработку
					t.Logf("failed to ACK message: %v", err)
				}
			} else {
				if err := msg.Nack(false, true); err != nil { // возвращаем в очередь
					t.Logf("failed to Nack message: %v", err)
				}
			}
			goto done
		case <-ctx.Done():
			goto done
		}
	}
done:

	assert.True(t, received, "notification was not received from RabbitMQ")

}

func checkListing(t *testing.T, url, expectedEventID string) {
	resp, err := http.Get(url)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d, body: %s", resp.StatusCode, string(body))
	}

	var events []Event
	err = json.NewDecoder(resp.Body).Decode(&events)
	require.NoError(t, err)

	found := false
	for _, ev := range events {
		if ev.ID == expectedEventID {
			found = true
			break
		}
	}
	assert.True(t, found, "event not found in listing: %s", url)
}
