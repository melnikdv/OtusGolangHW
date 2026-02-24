package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	event "github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/api"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage"
	"github.com/sirupsen/logrus"
)

type Server struct {
	event.UnimplementedEventServiceServer
	store  storage.Storage
	logger *logrus.Logger
	server *grpc.Server
	host   string
	port   int
}

func New(logger *logrus.Logger, store storage.Storage, host string, port int) *Server {
	return &Server{
		store:  store,
		logger: logger,
		host:   host,
		port:   port,
	}
}

func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.host, s.port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.server = grpc.NewServer(
		grpc.UnaryInterceptor(s.loggingInterceptor),
	)

	event.RegisterEventServiceServer(s.server, s)
	reflection.Register(s.server)

	s.logger.Infof("starting gRPC server on %s:%d", s.host, s.port)

	errCh := make(chan error, 1)
	go func() {
		if err := s.server.Serve(lis); err != nil {
			errCh <- fmt.Errorf("gRPC server failed: %w", err)
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		s.server.GracefulStop()
		return ctx.Err()
	}
}

func (s *Server) Stop(_ context.Context) error {
	s.logger.Info("shutting down gRPC server")
	s.server.GracefulStop()
	return nil
}

func (s *Server) loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)
	fields := logrus.Fields{
		"method":     info.FullMethod,
		"latency_ms": duration.Milliseconds(),
		"success":    err == nil,
	}

	if err != nil {
		fields["error"] = err.Error()
		s.logger.WithFields(fields).Error("gRPC request failed")
	} else {
		s.logger.WithFields(fields).Info("gRPC request handled")
	}

	return resp, err
}

// --- Реализация методов ---

func (s *Server) CreateEvent(_ context.Context, req *event.CreateEventRequest) (*event.CreateEventResponse, error) {
	ev := storage.Event{
		ID:           req.Event.Id,
		Title:        req.Event.Title,
		DateTime:     time.Unix(0, req.Event.StartTime),
		Duration:     (req.Event.EndTime - req.Event.StartTime) / 1e9,
		Description:  req.Event.Description,
		UserID:       req.Event.UserId,
		NotifyBefore: req.Event.NotifyBefore / 1e9,
	}

	if err := s.store.Add(ev); err != nil {
		return nil, s.handleError(err)
	}

	return &event.CreateEventResponse{Id: ev.ID}, nil
}

func (s *Server) UpdateEvent(_ context.Context, req *event.UpdateEventRequest) (*event.UpdateEventResponse, error) {
	ev := storage.Event{
		Title:        req.Event.Title,
		DateTime:     time.Unix(0, req.Event.StartTime),
		Duration:     (req.Event.EndTime - req.Event.StartTime) / 1e9,
		Description:  req.Event.Description,
		UserID:       req.Event.UserId,
		NotifyBefore: req.Event.NotifyBefore / 1e9,
	}

	if err := s.store.Update(req.Id, ev); err != nil {
		return nil, s.handleError(err)
	}

	return &event.UpdateEventResponse{}, nil
}

func (s *Server) DeleteEvent(_ context.Context, req *event.DeleteEventRequest) (*event.DeleteEventResponse, error) {
	if err := s.store.Delete(req.Id); err != nil {
		return nil, s.handleError(err)
	}
	return &event.DeleteEventResponse{}, nil
}

func (s *Server) ListEvents(_ context.Context, req *event.ListEventsRequest) (*event.ListEventsResponse, error) {
	var events []storage.Event
	var err error

	startDate := time.Unix(0, req.StartDate).Truncate(24 * time.Hour)

	switch req.Period {
	case event.ListEventsRequest_DAY:
		events, err = s.store.ListDay(startDate)
	case event.ListEventsRequest_WEEK:
		events, err = s.store.ListWeek(startDate)
	case event.ListEventsRequest_MONTH:
		events, err = s.store.ListMonth(startDate)
	default:
		return nil, status.Error(codes.InvalidArgument, "unknown period")
	}

	if err != nil {
		return nil, s.handleError(err)
	}

	apiEvents := make([]*event.Event, len(events))
	for i, ev := range events {
		endTime := ev.DateTime.Add(time.Duration(ev.Duration) * time.Second).UnixNano()
		apiEvents[i] = &event.Event{
			Id:           ev.ID,
			Title:        ev.Title,
			Description:  ev.Description,
			StartTime:    ev.DateTime.UnixNano(),
			EndTime:      endTime,
			UserId:       ev.UserID,
			NotifyBefore: ev.NotifyBefore * 1e9, // seconds → nanos
		}
	}

	return &event.ListEventsResponse{Events: apiEvents}, nil
}

func (s *Server) handleError(err error) error {
	if errors.Is(err, storage.ErrEventNotFound) {
		return status.Error(codes.NotFound, "event not found")
	}
	if errors.Is(err, storage.ErrDateBusy) {
		return status.Error(codes.AlreadyExists, "time slot is busy")
	}
	s.logger.WithError(err).Error("business logic error")
	return status.Error(codes.Internal, "internal server error")
}
