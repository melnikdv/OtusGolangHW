package hw06pipelineexecution

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	sleepPerStage = time.Millisecond * 100
	fault         = sleepPerStage / 2
)

func TestPipeline(t *testing.T) {
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("simple case", func(t *testing.T) {
		in := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()
		result := make([]string, 0, 5)
		start := time.Now()
		for s := range ExecutePipeline(in, nil, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Equal(t, []string{"102", "104", "106", "108", "110"}, result)
		require.Less(t,
			int64(elapsed),
			// ~0.8s for processing 5 values in 4 stages (100ms every) concurrently
			int64(sleepPerStage)*int64(len(stages)+len(data)-1)+int64(fault))
	})

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()
		result := make([]string, 0, 5)
		start := time.Now()
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		elapsed := time.Since(start)

		require.Len(t, result, 0)
		require.Less(t, int64(elapsed), int64(abortDur)+int64(fault))
	})
}

func TestAllStageStop(t *testing.T) {
	wg := sync.WaitGroup{}
	// Stage generator
	g := func(_ string, f func(v interface{}) interface{}) Stage {
		return func(in In) Out {
			out := make(Bi)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer close(out)
				for v := range in {
					time.Sleep(sleepPerStage)
					out <- f(v)
				}
			}()
			return out
		}
	}

	stages := []Stage{
		g("Dummy", func(v interface{}) interface{} { return v }),
		g("Multiplier (* 2)", func(v interface{}) interface{} { return v.(int) * 2 }),
		g("Adder (+ 100)", func(v interface{}) interface{} { return v.(int) + 100 }),
		g("Stringifier", func(v interface{}) interface{} { return strconv.Itoa(v.(int)) }),
	}

	t.Run("done case", func(t *testing.T) {
		in := make(Bi)
		done := make(Bi)
		data := []int{1, 2, 3, 4, 5}

		// Abort after 200ms
		abortDur := sleepPerStage * 2
		go func() {
			<-time.After(abortDur)
			close(done)
		}()

		go func() {
			for _, v := range data {
				in <- v
			}
			close(in)
		}()
		result := make([]string, 0, 5)
		for s := range ExecutePipeline(in, done, stages...) {
			result = append(result, s.(string))
		}
		wg.Wait()
		require.Len(t, result, 0)
	})
}

// Нет стейджей (pipeline = passthrough).
func TestPipelineNoStages(t *testing.T) {
	in := make(Bi)
	go func() {
		in <- 1
		in <- 2
		in <- 3
		close(in)
	}()
	result := make([]int, 0, 3)
	for v := range ExecutePipeline(in, nil) {
		result = append(result, v.(int))
	}

	require.Equal(t, []int{1, 2, 3}, result)
}

// done закрыт до старта пайплайна.
func TestPipelineDoneBeforeStart(t *testing.T) {
	in := make(Bi)
	done := make(Bi)
	close(done)

	go func() {
		in <- 1
		in <- 2
		close(in)
	}()
	result := make([]interface{}, 0, 2)
	for v := range ExecutePipeline(in, done,
		func(in In) Out {
			out := make(Bi)
			go func() {
				defer close(out)
				for v := range in {
					out <- v
				}
			}()
			return out
		},
	) {
		result = append(result, v)
	}

	require.Empty(t, result)
}

// Медленный потребитель (важно для backpressure).
func TestPipelineSlowConsumer(t *testing.T) {
	in := make(Bi)

	stage := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				out <- v.(int) * 2
			}
		}()
		return out
	}

	go func() {
		for i := 0; i < 5; i++ {
			in <- i
		}
		close(in)
	}()
	result := make([]int, 0, 5)
	for v := range ExecutePipeline(in, nil, stage) {
		time.Sleep(10 * time.Millisecond) // медленный consumer
		result = append(result, v.(int))
	}

	require.Equal(t, []int{0, 2, 4, 6, 8}, result)
}

// Стейдж не читает вход (защита от утечек).
func TestPipelineStageStopsReading(t *testing.T) {
	in := make(Bi)
	done := make(Bi)

	stage := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			// читаем только один элемент и выходим
			if v, ok := <-in; ok {
				out <- v
			}
		}()
		return out
	}

	go func() {
		in <- 1
		in <- 2
		in <- 3
		close(in)
	}()

	go func() {
		time.Sleep(20 * time.Millisecond)
		close(done)
	}()
	result := make([]int, 0, 1)
	for v := range ExecutePipeline(in, done, stage) {
		result = append(result, v.(int))
	}

	require.LessOrEqual(t, len(result), 1)
}

// Параллельность реально есть (не последовательное выполнение).
func TestPipelineStagesOverlap(t *testing.T) {
	in := make(Bi)

	var mu sync.Mutex
	active := 0
	maxActive := 0

	stage := func(in In) Out {
		out := make(Bi)
		go func() {
			defer close(out)
			for v := range in {
				mu.Lock()
				active++
				if active > maxActive {
					maxActive = active
				}
				mu.Unlock()

				time.Sleep(100 * time.Millisecond)

				mu.Lock()
				active--
				mu.Unlock()

				out <- v
			}
		}()
		return out
	}

	go func() {
		for i := 0; i < 3; i++ {
			in <- i
		}
		close(in)
	}()

	for range ExecutePipeline(in, nil, stage, stage, stage) {
		_ = struct{}{}
	}

	require.Greater(t, maxActive, 1)
}
