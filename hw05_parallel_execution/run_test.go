package hw05parallelexecution

import (
	"errors"
	"fmt"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestRun(t *testing.T) {
	defer goleak.VerifyNone(t)

	t.Run("if were errors in first M tasks, than finished not more N+M tasks", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32

		for i := 0; i < tasksCount; i++ {
			err := fmt.Errorf("error from task %d", i)
			tasks = append(tasks, func() error {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
				atomic.AddInt32(&runTasksCount, 1)
				return err
			})
		}

		workersCount := 10
		maxErrorsCount := 23
		err := Run(tasks, workersCount, maxErrorsCount)

		require.Truef(t, errors.Is(err, ErrErrorsLimitExceeded), "actual err - %v", err)
		require.LessOrEqual(t, runTasksCount, int32(workersCount+maxErrorsCount), "extra tasks were started")
	})

	t.Run("tasks without errors", func(t *testing.T) {
		tasksCount := 50
		tasks := make([]Task, 0, tasksCount)

		var runTasksCount int32
		var sumTime time.Duration

		for i := 0; i < tasksCount; i++ {
			taskSleep := time.Millisecond * time.Duration(rand.Intn(100))
			sumTime += taskSleep

			tasks = append(tasks, func() error {
				time.Sleep(taskSleep)
				atomic.AddInt32(&runTasksCount, 1)
				return nil
			})
		}

		workersCount := 5
		maxErrorsCount := 1

		start := time.Now()
		err := Run(tasks, workersCount, maxErrorsCount)
		elapsedTime := time.Since(start)
		require.NoError(t, err)

		require.Equal(t, int32(tasksCount), runTasksCount, "not all tasks were completed")
		require.LessOrEqual(t, int64(elapsedTime), int64(sumTime/2), "tasks were run sequentially?")
	})
}

// Проверяет выбранную трактовку: если m <= 0, то ошибки не ограничивают выполнение.
func TestRun_IgnoreErrorsWhenMNonPositive(t *testing.T) {
	defer goleak.VerifyNone(t)

	tasksCount := 20
	tasks := make([]Task, 0, tasksCount)

	var runTasksCount int32

	for i := 0; i < tasksCount; i++ {
		tasks = append(tasks, func() error {
			atomic.AddInt32(&runTasksCount, 1)
			return errors.New("some error")
		})
	}

	err := Run(tasks, 4, 0)

	require.NoError(t, err)
	require.Equal(t, int32(tasksCount), runTasksCount, "должны выполниться все задачи")
}

// n > len(tasks) — лишние воркеры не мешают.
func TestRun_WorkersMoreThanTasks(t *testing.T) {
	defer goleak.VerifyNone(t)

	tasksCount := 5
	tasks := make([]Task, 0, tasksCount)

	var runTasksCount int32

	for i := 0; i < tasksCount; i++ {
		tasks = append(tasks, func() error {
			atomic.AddInt32(&runTasksCount, 1)
			return nil
		})
	}

	err := Run(tasks, 10, 1)

	require.NoError(t, err)
	require.Equal(t, int32(tasksCount), runTasksCount)
}

// n == 1 — последовательное выполнение.
func TestRun_SingleWorker(t *testing.T) {
	defer goleak.VerifyNone(t)

	tasksCount := 10
	tasks := make([]Task, 0, tasksCount)

	var runTasksCount int32

	for i := 0; i < tasksCount; i++ {
		tasks = append(tasks, func() error {
			time.Sleep(10 * time.Millisecond)
			atomic.AddInt32(&runTasksCount, 1)
			return nil
		})
	}

	start := time.Now()
	err := Run(tasks, 1, 1)
	elapsed := time.Since(start)

	require.NoError(t, err)
	require.Equal(t, int32(tasksCount), runTasksCount)
	require.GreaterOrEqual(t, elapsed, 100*time.Millisecond, "задачи должны выполняться последовательно")
}

// Остановка строго при достижении лимита ошибок.
func TestRun_StopExactlyOnErrorLimit(t *testing.T) {
	defer goleak.VerifyNone(t)

	var runTasksCount int32

	tasks := []Task{
		func() error {
			atomic.AddInt32(&runTasksCount, 1)
			return errors.New("err1")
		},
		func() error {
			atomic.AddInt32(&runTasksCount, 1)
			return errors.New("err2")
		},
		func() error {
			atomic.AddInt32(&runTasksCount, 1)
			return nil
		},
		func() error {
			atomic.AddInt32(&runTasksCount, 1)
			return nil
		},
	}

	err := Run(tasks, 2, 2)

	require.ErrorIs(t, err, ErrErrorsLimitExceeded)
	require.LessOrEqual(t, runTasksCount, int32(4))
}

// Пустой список задач.
func TestRun_EmptyTasks(t *testing.T) {
	defer goleak.VerifyNone(t)

	err := Run(nil, 5, 1)
	require.NoError(t, err)
}
