package http

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage/inmemory"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestServer_Start_Stop(t *testing.T) {
	store := inmemory.New()
	logg := logrus.New()

	// Создаём слушатель на случайном порту
	l, err := net.Listen("tcp", "localhost:0")
	assert.NoError(t, err)
	defer func() { _ = l.Close() }()

	realAddr := l.Addr().String()

	srv := New(logg, store, "", 0)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		// Создаём mux отдельно
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("OK\n"))
		})

		// Применяем middleware
		srv.server.Handler = srv.loggingMiddleware(mux)

		// Запускаем сервер с существующим слушателем
		errCh <- srv.server.Serve(l)
	}()

	// Ждём запуска
	time.Sleep(50 * time.Millisecond)

	// Делаем запрос
	resp, err := http.Get("http://" + realAddr + "/")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := make([]byte, 3)
	n, _ := resp.Body.Read(body)
	assert.Equal(t, "OK\n", string(body[:n]))
	_ = resp.Body.Close()

	// Останавливаем
	cancel()
	_ = srv.Stop(ctx)

	select {
	case err := <-errCh:
		if err != nil && err != http.ErrServerClosed {
			t.Fatalf("Server failed: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Server did not stop in time")
	}
}

// Тест middleware через httptest
func TestLoggingMiddleware(t *testing.T) {
	logg := &logrus.Logger{
		Out: &strings.Builder{},
		Formatter: &logrus.TextFormatter{
			DisableTimestamp: true,
			DisableColors:    true,
		},
		Level: logrus.InfoLevel,
	}

	// Перехватываем вывод
	var buf strings.Builder
	logg.SetOutput(&buf)

	store := inmemory.New()
	srv := New(logg, store, "localhost", 0)

	// Создаём запрос напрямую через middleware
	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	middleware := srv.loggingMiddleware(handler)
	middleware.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTeapot, rec.Code)

	// Проверяем, что лог содержит нужные поля
	logStr := buf.String()
	assert.Contains(t, logStr, "GET")
	assert.Contains(t, logStr, "/test")
	assert.Contains(t, logStr, "418") // Teapot
}
