#!/bin/bash

env GOOS=linux \
go build -mod=mod -ldflags "\
  -s -w "\
  -o ./bin/feed \
  ./src/txt