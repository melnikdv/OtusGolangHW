package server

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type Server struct {
	logger *logrus.Logger
	server *http.Server
}

func New(logger *logrus.Logger, host string, port int) *Server {
	return &Server{
		logger: logger,
		server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", host, port),
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
		},
	}
}

func (s *Server) Start() error {
	// Простой "hello-world" handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, err := w.Write([]byte("Hello from Calendar Service!\n"))
		if err != nil {
			// Запишем в лог, но не можем повлиять на ответ — клиент уже отключился
			s.logger.WithError(err).Warn("Failed to write response")
		}
	})

	// Оборачиваем в middleware
	s.server.Handler = s.loggingMiddleware(handler)

	s.logger.Infof("Starting HTTP server on %s", s.server.Addr)
	return s.server.ListenAndServe()
}

func (s *Server) Stop() error {
	return s.server.Close()
}

// loggingMiddleware — соответсвует требованиям ТЗ.
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Оборачиваем ResponseWriter, чтобы перехватить код ответа
		lw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(lw, r)

		duration := time.Since(start)

		// Формат лога: IP [timestamp] METHOD PATH PROTO STATUS DURATION_MS "USER_AGENT"
		s.logger.Infof(
			`%s [%s] %s %s %s %d %d "%s"`,
			r.RemoteAddr,
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.URL.String(),
			r.Proto,
			lw.statusCode,
			duration.Milliseconds(),
			r.UserAgent(),
		)
	})
}

// loggingResponseWriter — позволяет получить реальный код ответа.
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
