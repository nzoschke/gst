package widgets

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/empty"
	widgets "github.com/nzoschke/gst/gen/go/widgets/v0"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Server implements the widgets/v0 interface
type Server struct{}

var _ widgets.WidgetsServer = (*Server)(nil) // assert interface

// BatchGet Widgets by names
func (s *Server) BatchGet(ctx context.Context, r *widgets.BatchGetRequest) (*widgets.BatchGetResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

// Create Widget
func (s *Server) Create(ctx context.Context, r *widgets.CreateRequest) (*widgets.Widget, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

// Delete Widget
func (s *Server) Delete(ctx context.Context, r *widgets.DeleteRequest) (*empty.Empty, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

// Get Widgets
func (s *Server) Get(ctx context.Context, r *widgets.GetRequest) (*widgets.Widget, error) {
	return &widgets.Widget{
		Parent:      "users/foo",
		Name:        r.Name,
		DisplayName: fmt.Sprintf("My %s widget", r.Name),
		CreateTime:  ptypes.TimestampNow(),
	}, nil
}

// List Widgets with pagination
func (s *Server) List(ctx context.Context, r *widgets.ListRequest) (*widgets.ListResponse, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}

// Update Widget with update_mask
func (s *Server) Update(ctx context.Context, r *widgets.UpdateRequest) (*widgets.Widget, error) {
	return nil, status.Error(codes.Unimplemented, "unimplemented")
}
