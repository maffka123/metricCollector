package server

import (
	"net"
	"testing"

	"google.golang.org/grpc/codes"

	"context"
	"log"

	"github.com/maffka123/metricCollector/internal/server/config"
	"go.uber.org/zap"

	pb "github.com/maffka123/metricCollector/internal/proto"
	"github.com/maffka123/metricCollector/internal/storage"
	"github.com/stretchr/testify/assert"
	grpc "google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

//http://www.inanzzz.com/index.php/post/w9qr/unit-testing-golang-grpc-client-and-server-application-with-bufconn-package
func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)
	logger, _ := zap.NewDevelopmentConfig().Build()
	cfg := config.Config{}
	db := storage.Connect(&cfg, logger)

	server := grpc.NewServer()

	pb.RegisterMetricsServer(server, &MetricsServer{db: db, log: logger})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestAddMetricCounter(t *testing.T) {
	ctx := context.Background()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewMetricsClient(conn)

	tests := []struct {
		name    string
		amount  int64
		res     *pb.StdResponse
		errCode codes.Code
		errMsg  string
	}{
		{
			name:    "add counter",
			amount:  2,
			res:     &pb.StdResponse{},
			errCode: codes.InvalidArgument,
			errMsg:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &pb.Metric{Id: "counter1", Delta: tt.amount}

			response, err := client.AddMetricCounter(ctx, request)

			assert.NotNil(t, response)
			assert.Equal(t, response.Error, tt.res.Error)
			if err != nil {
				assert.Equal(t, err.Error(), tt.errMsg)
			}

		})
	}
}
