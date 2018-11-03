# gst -- gRPC Service Template

Quickstart:

```shell
$ docker build -f Dockerfile-prototool -t prototool .
$ make dev
$ curl localhost/v0/widgets/blue
```

```json
{
 "parent": "users/foo",
 "name": "widgets/blue",
 "display_name": "My widgets/blue widget",
 "create_time": "2018-11-03T01:06:09.832085600Z"
}
```

## Copy boilerplate config

| File / Directory                   | Function                      |
|------------------------------------|-------------------------------|
| go.mod                             | Go module config              |
| Makefile                           | Build, test, dev commands     |
| bin/prototool.sh                   | `prototool` build script      |
| proto/widgets/v0/widgets.proto     | 1st party .protos             |
| proto/prototool.yaml               | 1st party prototool config    |
| proto_ext/github.com/              | 3rd party .protos             |
| proto_ext/prototool.yaml           | 3rd party prototool config    | 
| config/docker/compose.yaml         | Dev service definitions       |
| config/docker/Dockerfile           | Service runtime environment   |
| config/docker/Dockerfile-prototool | `prototool` build environment |
| config/envoy/proxy.yaml            | API gateway config            |
| config/envoy/sidecar.yaml          | Service sidecar config        |

## Write .proto file

```shell
$ PROTO=widgets/v0/widgets.proto make create
```

```proto
syntax = "proto3";

package gst.widgets.v0;

option go_package = "v0pb";
option java_multiple_files = true;
option java_outer_classname = "WidgetsProto";
option java_package = "com.gst.widgets.v0";

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";
import "google/protobuf/field_mask.proto";
import "google/protobuf/timestamp.proto";
import "validate/validate.proto";

service Widgets {
  rpc Get(GetRequest) returns (Widget) {
    option (google.api.http) = {
      get: "/v3/{name=users/*}"
    };
  }
  rpc Create(CreateRequest) returns (Widget);
  rpc Update(UpdateRequest) returns (Widget);
  rpc Delete(DeleteRequest) returns (google.protobuf.Empty);
  rpc List(ListRequest) returns (ListResponse);
  rpc BatchGet(BatchGetRequest) returns (BatchGetResponse);
}

message Widget {
  string parent = 1;
  string name = 2;
  string display_name = 3;
  google.protobuf.Timestamp create_time = 4;
}

message GetRequest {
  string name = 1 [
    (validate.rules).string = {
      pattern: "^users/[a-z0-9._-]+$"
      max_bytes: 512
    }
  ];
}

message CreateRequest {
  string parent = 1;
  string id = 2;
  Widget widget = 3;
}

message UpdateRequest {
  Widget widget = 1;
  google.protobuf.FieldMask update_mask = 2;
}

message DeleteRequest {
  string name = 1;
}

message ListRequest {
  string parent = 1;
  int32 page_size = 2;
  string page_token = 3;
}

message ListResponse {
  repeated Widget widgets = 1;
  string next_page_token = 2;
}

message BatchGetRequest {
  string parent = 1;
  repeated string names = 2;
}

message BatchGetResponse {
  repeated Widget widgets = 1;
}
```
> From proto/widgets/v0/widgets.proto

## Build service interface

```shell
$ make build
```

## Write server and service interface

```go
package main

import (
	"fmt"
	"net"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_validator "github.com/grpc-ecosystem/go-grpc-middleware/validator"
	widgets "github.com/nzoschke/gst/gen/go/widgets/v0"
	swidgets "github.com/nzoschke/gst/server/widgets/v0"
	"github.com/segmentio/conf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type config struct {
	Port int `conf:"p" help:"Port to listen"`
}

func main() {
	config := config{
		Port: 8000,
	}
	conf.Load(&config)

	if err := serve(config); err != nil {
		panic(err)
	}
}

func serve(config config) error {
	s := grpc.NewServer(
		grpc.UnaryInterceptor(
			grpc_middleware.ChainUnaryServer(
				grpc_validator.UnaryServerInterceptor(),
			),
		),
	)
	widgets.RegisterWidgetsServer(s, &swidgets.Server{})
	reflection.Register(s)

	l, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", config.Port))
	if err != nil {
		return err
	}

	fmt.Printf("listening on :%d\n", config.Port)
	return s.Serve(l)
}
```
> From bin/widgets-v0/main.go

```go
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
```
> From server/widgets/v0/widgets.go

## Run gRPC service and sidecar

```shell
$ make dev
```

Make a request with `prototool` to port 8000:

```shell
$ prototool grpc proto              \
--address 0.0.0.0:8000              \
--method gst.widgets.v0.Widgets/Get \
--data '{
    "name": "widgets/red"
  }'
```

```json
{
  "parent": "users/foo",
  "name": "widgets/red",
  "displayName": "My widgets/red widget",
  "createTime": "2018-11-03T01:05:39.841369700Z"
}
```

## Add API gateway config

```yaml
staticResources:
  listeners:
  - address:
      socketAddress:
        address: 0.0.0.0
        portValue: 10000
    filterChains:
    - filters:
      - config:
          httpFilters:
          - name: envoy.grpc_web
          - name: envoy.grpc_json_transcoder
            config:
              print_options:
                add_whitespace: true
                preserve_proto_field_names: true
                always_print_primitive_fields: true
              proto_descriptor: /etc/pb/widgets/v0/widgets.pb
              services:
                - gst.widgets.v0.Widgets
          - name: envoy.router
          routeConfig:
            name: local
            virtualHosts:
            - domains:
              - '*'
              name: local
              routes:
              - match:
                  prefix: /gst.widgets.v0.Widgets
                route:
                  cluster: widgets-v0
        name: envoy.http_connection_manager
    name: ingress
```
> From config/envoy/proxy.yaml

## Run API gateway

```shell
$ make dev
```

Make a request with `curl` to port 80:

```shell
$ curl localhost/v0/widgets/blue
```

```json
{
 "parent": "users/foo",
 "name": "widgets/blue",
 "display_name": "My widgets/blue widget",
 "create_time": "2018-11-03T01:06:09.832085600Z"
}
```