package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run запускает задачи tasks в n параллельных горутинах.
// Работа прекращается при получении m ошибок (если m > 0).
func Run(tasks []Task, n, m int) error {
	// Канал с задачами для воркеров
	tasksCh := make(chan Task)

	// Счётчик ошибок
	var errorsCount int32

	// WaitGroup для ожидания завершения всех воркеров
	var wg sync.WaitGroup

	// Функция воркера
	worker := func() {
		defer wg.Done()
		for task := range tasksCh {
			// Если лимит ошибок достигнут — прекращаем работу
			if m > 0 && atomic.LoadInt32(&errorsCount) >= int32(m) {
				return
			}

			// Выполняем задачу
			if err := task(); err != nil {
				// Увеличиваем счётчик ошибок
				if m > 0 {
					atomic.AddInt32(&errorsCount, 1)
				}
			}
		}
	}

	// Запускаем n воркеров
	wg.Add(n)
	for i := 0; i < n; i++ {
		go worker()
	}

	// Отправляем задачи воркерам
	for _, task := range tasks {
		// Если лимит ошибок достигнут — прекращаем отправку задач
		if m > 0 && atomic.LoadInt32(&errorsCount) >= int32(m) {
			break
		}
		tasksCh <- task
	}

	// Закрываем канал задач — сигнал воркерам завершаться
	close(tasksCh)

	// Ждём завершения всех воркеров
	wg.Wait()

	// Если лимит ошибок превышен — возвращаем ошибку
	if m > 0 && atomic.LoadInt32(&errorsCount) >= int32(m) {
		return ErrErrorsLimitExceeded
	}

	return nil
}
