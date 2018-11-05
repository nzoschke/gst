# gST - gRPC Service Template

## Quickstart

First we need to:

* build a Docker image with proto tools
* generate go, js and ts files
* generate go mock client / servers
* build linux server binaries

```shell
$ make build
```

Now we can start the gRPC servers and API gateway:

```shell
$ make dev
```

Then we can make a gRPC request directly to a service:

```shell
$ prototool grpc proto              \
--address 0.0.0.0:8000              \
--method gst.widgets.v0.Widgets/Get \
--data '{
    "name": "widgets/red"
  }'
```

Example response:

```json
{
  "parent": "users/foo",
  "name": "widgets/red",
  "displayName": "My widgets/red widget",
  "createTime": "2018-11-05T02:36:42.599780200Z"
}
```

And we can make a REST request to the API Gateway:

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

## Boostrapping a service

First we start with boilerplate config:

| File / Directory                   | Function                          |
|------------------------------------|-----------------------------------|
| bin/protogen.sh                    | `prototool` build script          |
| config/envoy/proxy.yaml            | API gateway config                |
| config/envoy/sidecar.yaml          | Service sidecar config            |
| proto/prototool.yaml               | 1st party protos and build config |
| proto_ext/prototool.yaml           | 3rd party protos and build config |
| docker-compose.yaml                | Dev service definitions           |
| Dockerfile                         | Service runtime environment       |
| Dockerfile-protogen                | `prototool` build environment     |
| go.mod                             | Go module config                  |
| Makefile                           | Build, test, dev commands         |

To this we will add our custom service definition and business logic:

| File / Directory                  | Function               |
|-----------------------------------|------------------------|
| cmd/widgets-v0/main.go            | gRPC server program    |
| proto/widgets/v0/widgets.proto    | Service definition     |
| server/widgets/v0/widgets_test.go | Service tests          |
| server/widgets/v0/widgets.go      | Service implementation |

### Service definition

First we start with boilerplate .proto headers:

```shell
$ PROTO=widgets/v0/widgets.proto make create
```

To this we add our service name, methods, request and response methods. This should implement the standard methods from the API style guide:

<details>
<summary>See an example (proto/widgets/v0/widgets.proto)[proto/widgets/v0/widgets.proto] file...</summary>

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
      get: "/v0/{name=users/*/widgets/*}"
    };
  }
  rpc Create(CreateRequest) returns (Widget) {
    option (google.api.http) = {
      post: "/v0/{parent=users/*}/widgets"
      body: "*"
    };
  }
  rpc Update(UpdateRequest) returns (Widget) {
    option (google.api.http) = {
      patch: "/v0/{widget.name=users/*/widgets/*}"
      body: "*"
    };
  }
  rpc Delete(DeleteRequest) returns (google.protobuf.Empty) {
    option (google.api.http) = {
      delete: "/v0/{name=users/*/widgets/*}"
    };
  }
  rpc List(ListRequest) returns (ListResponse) {
    option (google.api.http) = {
      get: "/v0/{parent=users/*}/widgets"
    };
  }
  rpc BatchGet(BatchGetRequest) returns (BatchGetResponse) {
    option (google.api.http) = {
      get: "/v0/{parent=users/*}/widgets:batchGet"
    };
  }
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
      pattern: "^users/[a-z0-9._-]+/widgets/[a-z0-9._-]+$"
      max_bytes: 512
    }
  ];
}

message CreateRequest {
  string parent = 1 [
    (validate.rules).string = {
      pattern: "^users/[a-z0-9._-]+$"
      max_bytes: 512
    }
  ];
  string id = 2 [
    (validate.rules).string = {
      pattern: "^[a-z0-9._-]+$"
      max_bytes: 512
    }
  ];
  Widget widget = 3 [(validate.rules).message.required = true];
}

message UpdateRequest {
  Widget widget = 1 [(validate.rules).message.required = true];
  google.protobuf.FieldMask update_mask = 2 [(validate.rules).message.required = true];
}

message DeleteRequest {
  string name = 1 [
    (validate.rules).string = {
      pattern: "^users/[a-z0-9._-]+/widgets/[a-z0-9._-]+$"
      max_bytes: 512
    }
  ];
}

message ListRequest {
  string parent = 1 [
    (validate.rules).string = {
      pattern: "^users/[a-z0-9._-]+$"
      max_bytes: 512
    }
  ];
  int32 page_size = 2;
  string page_token = 3;
}

message ListResponse {
  repeated Widget widgets = 1;
  string next_page_token = 2;
}

message BatchGetRequest {
  string parent = 1 [
    (validate.rules).string = {
      pattern: "^users/[a-z0-9._-]+$"
      max_bytes: 512
    }
  ];
  repeated string names = 2;
}

message BatchGetResponse {
  repeated Widget widgets = 1;
}
```
</details>&nbsp;

### Service implementation

Now we need to build the service business logic. First generate the go gRPC service interface:

```shell
$ make gen
```

Next write a package that implements the gRPC service interface:

<details>
<summary>See an example (server/widgets/v0/widgets.go)[server/widgets/v0/widgets.go] file...</summary>

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
</details>&nbsp;

Then write a gRPC server program that uses the grpc_validator middleware:

<details>
<summary>See an example (bin/widgets-v0/main.go)[bin/widgets-v0/main.go] file...</summary>


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
</details>&nbsp;

### Service proxy config

Now we need to configure access to the service. First add a docker-compose service that exposes our gRPC server. Note that we are exposing `10000`, the envoy proxy sidecar listener:

```yaml
services:
  widgets-v0:
    build: .
    command: ["widgets-v0"]
    environment:
      - ENVOY_CONFIG=/etc/envoy/sidecar.yaml
    ports:
      - "8000:10000"  # map envoy ingress
```

Now you can send a gRPC request to port `8000`:

```shell
$ make dev
```

```shell
$ prototool grpc proto              \
--address 0.0.0.0:8000              \
--method gst.widgets.v0.Widgets/Get \
--data '{
    "name": "users/foo/widgets/red"
  }'
```

Example response:

```json
{
  "parent": "users/foo",
  "name": "users/foo/widgets/red",
  "displayName": "My users/foo/widgets/red widget",
  "createTime": "2018-11-05T18:13:52.653833300Z"
}
```

Finally add Envoy proxy configuration that registers the service in the REST gateway.

<details>
<summary>See an example (config/envoy/proxy.yaml)[config/envoy/proxy.yaml] file...</summary>


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
</details>&nbsp;

Now you can send a REST request to port `80`:

```shell
$ make dev
```

```shell
$ curl localhost/v0/users/foo/widgets/blue
```

Example response:

```json
{
 "parent": "users/foo",
 "name": "users/foo/widgets/blue",
 "display_name": "My users/foo/widgets/blue widget",
 "create_time": "2018-11-05T18:14:23.269033300Z"
}
```

### Clients

Now we can use the generated clients to access the service. Provided is an example typescript gRPC-web client"

```shell
$ node examples/client_ts/index.js
```

Example response:

```js
{ parent: 'users/foo',
  name: 'users/foo/widgets/blue',
  displayName: 'My users/foo/widgets/blue widget',
  createTime: { seconds: 1541442585, nanos: 203350900 } }
```