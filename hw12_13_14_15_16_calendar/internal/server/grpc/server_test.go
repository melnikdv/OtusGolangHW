package grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	event "github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/api"
	"github.com/melnikdv/OtusGolangHW/hw12_13_14_15_16_calendar/internal/storage/inmemory"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const bufSize = 1024 * 1024

func bufDialer(lis *bufconn.Listener) func(context.Context, string) (net.Conn, error) {
	return func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}
}

func TestGRPC_CreateAndListEvent(t *testing.T) {
	lis := bufconn.Listen(bufSize)
	defer func() { _ = lis.Close() }()

	store := inmemory.New()
	logg := logrus.New()
	srv := New(logg, store, "", 0)

	grpcServer := grpc.NewServer()
	event.RegisterEventServiceServer(grpcServer, srv)
	go func() { _ = grpcServer.Serve(lis) }()
	defer grpcServer.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(lis)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer func() { _ = conn.Close() }()

	client := event.NewEventServiceClient(conn)

	start := time.Date(2026, 2, 10, 10, 0, 0, 0, time.UTC).UnixNano()
	end := start + 3600*1e9
	_, err = client.CreateEvent(ctx, &event.CreateEventRequest{
		Event: &event.Event{
			Id:        "1",
			Title:     "gRPC Test",
			StartTime: start,
			EndTime:   end,
			UserId:    "user1",
		},
	})
	assert.NoError(t, err)

	resp, err := client.ListEvents(ctx, &event.ListEventsRequest{
		Period:    event.ListEventsRequest_DAY,
		StartDate: start,
	})
	assert.NoError(t, err)
	assert.Len(t, resp.Events, 1)
	assert.Equal(t, "gRPC Test", resp.Events[0].Title)
}

func TestGRPC_DeleteNonExistentEvent(t *testing.T) {
	lis := bufconn.Listen(bufSize)
	defer func() { _ = lis.Close() }()

	store := inmemory.New()
	logg := logrus.New()
	srv := New(logg, store, "", 0)

	grpcServer := grpc.NewServer()
	event.RegisterEventServiceServer(grpcServer, srv)
	go func() { _ = grpcServer.Serve(lis) }()
	defer grpcServer.Stop()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(bufDialer(lis)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	assert.NoError(t, err)
	defer func() { _ = conn.Close() }()

	client := event.NewEventServiceClient(conn)

	_, err = client.DeleteEvent(ctx, &event.DeleteEventRequest{Id: "nonexistent"})
	assert.Error(t, err)
	st, _ := status.FromError(err)
	assert.Equal(t, codes.NotFound, st.Code())
}
