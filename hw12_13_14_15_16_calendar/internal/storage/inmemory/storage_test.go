package inmemory

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryStorage(t *testing.T) {
	s := New()
	now := time.Now()
	event := storage.Event{
		ID:       "1",
		Title:    "Test",
		DateTime: now,
		Duration: 3600,
		UserID:   "user1",
	}

	// Add
	assert.NoError(t, s.Add(event))
	assert.ErrorIs(t, s.Add(event), storage.ErrDateBusy)

	// List
	events, _ := s.ListDay(now)
	assert.Len(t, events, 1)

	// Update
	event.Title = "Updated"
	assert.NoError(t, s.Update("1", event))

	// Delete
	assert.NoError(t, s.Delete("1"))
	assert.ErrorIs(t, s.Delete("1"), storage.ErrEventNotFound)
}

func TestInMemoryStorage_DateBusyPerUser(t *testing.T) {
	s := New()
	now := time.Now()

	event1 := storage.Event{
		ID:       "1",
		UserID:   "user1",
		DateTime: now,
		Title:    "Event 1",
		Duration: 3600,
	}
	event2 := storage.Event{
		ID:       "2",
		UserID:   "user1", // ← тот же пользователь
		DateTime: now,     // ← то же время
		Title:    "Event 2",
		Duration: 3600,
	}
	event3 := storage.Event{
		ID:       "3",
		UserID:   "user2", // ← другой пользователь
		DateTime: now,     // ← то же время — должно быть OK
		Title:    "Event 3",
		Duration: 3600,
	}

	assert.NoError(t, s.Add(event1))
	assert.ErrorIs(t, s.Add(event2), storage.ErrDateBusy) // ← ожидаем ошибку
	assert.NoError(t, s.Add(event3))                      // ← разрешено
}

func TestInMemoryStorage_ListWeekAndMonth(t *testing.T) {
	s := New()
	now := time.Now()

	// Событие сегодня
	eventToday := storage.Event{ID: "1", UserID: "user1", DateTime: now, Duration: 3600}
	// Событие через 3 дня (в той же неделе)
	eventIn3Days := storage.Event{ID: "2", UserID: "user1", DateTime: now.Add(3 * 24 * time.Hour), Duration: 3600}
	// Событие через 10 дней (та же неделя? зависит от now, но точно в том же месяце)
	eventIn10Days := storage.Event{ID: "3", UserID: "user1", DateTime: now.Add(10 * 24 * time.Hour), Duration: 3600}
	// Событие в следующем месяце
	nextMonth := now.AddDate(0, 1, 0)
	eventNextMonth := storage.Event{ID: "4", UserID: "user1", DateTime: nextMonth, Duration: 3600}

	assert.NoError(t, s.Add(eventToday))
	assert.NoError(t, s.Add(eventIn3Days))
	assert.NoError(t, s.Add(eventIn10Days))
	assert.NoError(t, s.Add(eventNextMonth))

	// ListWeek
	weekEvents, err := s.ListWeek(now)
	assert.NoError(t, err)
	assert.Len(t, weekEvents, 2) // today + in 3 days

	// ListMonth
	monthEvents, err := s.ListMonth(now)
	assert.NoError(t, err)
	assert.Len(t, monthEvents, 3) // все, кроме nextMonth
}

func TestInMemoryStorage_Concurrency(t *testing.T) {
	s := New()
	now := time.Now()
	const numWorkers = 10
	const eventsPerWorker = 100

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < eventsPerWorker; j++ {
				event := storage.Event{
					ID:       fmt.Sprintf("worker-%d-event-%d", workerID, j),
					UserID:   fmt.Sprintf("user-%d", workerID),
					DateTime: now.Add(time.Duration(j) * time.Hour),
					Duration: 3600,
					Title:    fmt.Sprintf("Event %d-%d", workerID, j),
				}
				err := s.Add(event)
				// Должно быть без ошибок, так как у каждого свой UserID и время
				assert.NoError(t, err)
			}
		}(i)
	}

	wg.Wait()

	// Проверим общее количество
	events, _ := s.ListMonth(now)
	assert.Len(t, events, numWorkers*eventsPerWorker)
}
