#!/bin/bash
set -e -o pipefail
PWD=$(pwd)

prototool format -w proto
prototool generate proto
prototool generate proto_ext

# Piggy back on `prototool compile` to run custom commands for every .proto input
#
# Example output from `prototool compile --dry-run`:
#
# /root/.cache/prototool/Linux/x86_64/protobuf/3.6.1/bin/protoc \
#   -I /in/vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
#   -I /in/vendor/github.com/lyft/protoc-gen-validate \
#   -I /root/.cache/prototool/Linux/x86_64/protobuf/3.6.1/include \
#   -I /in/proto \
#   -o /dev/null \
#   /in/proto/users/v1/users.proto

prototool compile proto --dry-run | while read x; do
    # get last arg
    IN=$(echo $x | grep -oE "[^ ]+$")      # /in/proto/users/v1/users.proto

    # get arg components
    IN_FILE=${IN/$PWD\/proto/}             # users/v1/users.proto
    IN_DIR=$(dirname $IN_FILE)             # users/v1

    # make dir and compile .pb
    OUT_DIR=gen/pb/${IN_DIR}               # gen/pb/users/v1
    mkdir -p $OUT_DIR

    OUT_FILE=gen/pb/${IN_FILE/.proto/.pb}  # gen/pb/users/v1/users.pb
    CMD=${x/\/dev\/null/$OUT_FILE}         # replace /dev/null with output
    $CMD --include_imports
done
