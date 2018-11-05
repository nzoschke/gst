# build on .go file changes
CMD = $(wildcard cmd/*)
BIN = $(CMD:cmd/%=bin/linux_amd64/%)
$(BIN): bin/linux_amd64/%: cmd/%/main.go $(shell find . -name '*.go')
	GOOS=linux GOARCH=amd64 go build -o $@ $<

# generate on .proto file changes
PROTO = $(wildcard proto/*/*/*.proto)
PBGO = $(PROTO:proto/%.proto=gen/go/%.pb.go)
$(PBGO): gen/go/%.pb.go: bin/protogen.sh proto/prototool.yaml proto/%.proto proto_ext/prototool.yaml
	docker build -f Dockerfile-protogen -t protogen .
	docker run -v $(PWD):/in -v $(PWD)/bin/protogen.sh:/bin/protogen.sh protogen /bin/protogen.sh

build: gen $(BIN)

clean:
	rm -rf bin/linux_amd64/*
	rm -rf gen

create:
	mkdir -p proto/$(dir $(PROTO))
	docker run -v $(PWD):/in prototool prototool create proto/$(PROTO)
	cat proto/$(PROTO)

dev: build
	docker-compose up --abort-on-container-exit

gen: $(PBGO)

list:
	@grep -o "^[a-z-]*:" Makefile

test:
	go test -v ./...