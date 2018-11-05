package widgets_test

import (
	"context"
	"fmt"
	"net"
	"testing"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	widgets "github.com/nzoschke/gst/gen/go/widgets/v0"
	"github.com/nzoschke/gst/internal/clock"
	swidgets "github.com/nzoschke/gst/server/widgets/v0"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var Clock = clock.TestClock{}

func init() {
	swidgets.Clock = Clock
}

func server(t *testing.T, srv widgets.WidgetsServer) (net.Listener, *grpc.Server, *grpc.ClientConn) {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_validator.UnaryServerInterceptor(),
			),
		),
	)
	widgets.RegisterWidgetsServer(s, srv)

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", 0))
	assert.NoError(t, err)

	c, err := grpc.Dial(l.Addr().String(), grpc.WithInsecure())
	assert.NoError(t, err)

	return l, s, c
}

// TestGet demonstrates a real client and server Get, GetRequest and GetResponse
func TestGet(t *testing.T) {
	srv := &swidgets.Server{}
	l, s, conn := server(t, srv)
	go s.Serve(l)
	defer s.GracefulStop()
	c := widgets.NewWidgetsClient(conn)

	_, err := c.Get(context.Background(), &widgets.GetRequest{
		Name: "invalid",
	})
	assert.Error(t, err)
	serr := status.Convert(err)
	assert.Equal(t, codes.InvalidArgument, serr.Code())
	assert.Equal(t, `invalid GetRequest.Name: value does not match regex pattern "^widgets/[a-z0-9._-]+$"`, serr.Message())

	w, err := c.Get(context.Background(), &widgets.GetRequest{
		Name: "widgets/red",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, &widgets.Widget{
		Parent:      "users/foo",
		Name:        "widgets/red",
		DisplayName: "My widgets/red widget",
		CreateTime:  Clock.TimestampNow(),
	}, w)
}

// TestMockClient demonstrates a client mock
func TestMockClient(t *testing.T) {
	c := widgets.MockWidgetsClient{}
	c.On(
		"Get",
		mock.Anything,
		&widgets.GetRequest{
			Name: "widgets/red",
		},
	).Return(
		nil,
		status.Error(codes.Unavailable, "unavailable"),
	)
	_, err := c.Get(context.Background(), &widgets.GetRequest{
		Name: "widgets/red",
	})
	assert.EqualError(t, err, "rpc error: code = Unavailable desc = unavailable")
}

// TestMockServer demonstrates a server mock
func TestMockServer(t *testing.T) {
	srv := &widgets.MockWidgetsServer{}
	l, s, conn := server(t, srv)
	go s.Serve(l)
	defer s.GracefulStop()
	c := widgets.NewWidgetsClient(conn)

	srv.On(
		"List",
		mock.Anything,
		&widgets.ListRequest{
			Parent: "users/foo",
		},
	).Return(
		&widgets.ListResponse{
			Widgets: []*widgets.Widget{
				&widgets.Widget{
					Parent:      "users/foo",
					Name:        "users/foo/widgets/red",
					DisplayName: "Red",
					CreateTime:  Clock.TimestampNow(),
				},
			},
		},
		nil,
	)

	w, err := c.List(context.Background(), &widgets.ListRequest{
		Parent: "users/foo",
	})
	assert.NoError(t, err)
	assert.EqualValues(t, &widgets.ListResponse{
		Widgets: []*widgets.Widget{
			&widgets.Widget{
				Parent:      "users/foo",
				Name:        "users/foo/widgets/red",
				DisplayName: "Red",
				CreateTime:  Clock.TimestampNow(),
			},
		},
	}, w)

}
