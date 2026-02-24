package http

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/sirupsen/logrus"
)

type Server struct {
	logger *logrus.Logger
	store  storage.Storage
	server *http.Server
}

func New(logger *logrus.Logger, store storage.Storage, host string, port int) *Server {
	return &Server{
		logger: logger,
		store:  store,
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", host, port),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		if _, err := w.Write([]byte("OK\n")); err != nil {
			s.logger.WithError(err).Warn("failed to write response")
		}
	})

	// Подключаем хендлеры
	h := &handler{store: s.store, logger: s.logger}
	mux.HandleFunc("POST /events", h.createEvent)
	mux.HandleFunc("PUT /events/{id}", h.updateEvent)
	mux.HandleFunc("DELETE /events/{id}", h.deleteEvent)
	mux.HandleFunc("GET /events/day", h.listDay)
	mux.HandleFunc("GET /events/week", h.listWeek)
	mux.HandleFunc("GET /events/month", h.listMonth)

	s.server.Handler = s.loggingMiddleware(mux)

	errCh := make(chan error, 1)
	go func() {
		s.logger.Infof("starting HTTP server on %s", s.server.Addr)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("shutting down HTTP server")
	return s.server.Shutdown(ctx)
}

// loggingMiddleware — как у вас
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(lrw, r)
		duration := time.Since(start)
		s.logger.Infof(
			`%s [%s] %s %s %s %d %d "%s"`,
			r.RemoteAddr,
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.URL.String(),
			r.Proto,
			lrw.statusCode,
			duration.Milliseconds(),
			r.UserAgent(),
		)
	})
}

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
