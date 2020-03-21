#!/bin/bash
go build ./...
go build -tags dynamicplugintesting -buildmode=plugin \
    -o internal/dynamicplugintesting/dynfoo/dynfooplug.so \
    ./internal/dynamicplugintesting/dynfoo
go test -v -timeout 30s ./... -cover