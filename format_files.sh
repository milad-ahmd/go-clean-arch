#!/bin/bash

echo "Formatting pkg/errors/errors.go"
gofmt -s -w pkg/errors/errors.go

echo "Formatting internal/domain/errors.go"
gofmt -s -w internal/domain/errors.go

echo "Formatting pkg/response/response.go"
gofmt -s -w pkg/response/response.go

echo "Formatting internal/usecase/user_usecase.go"
gofmt -s -w internal/usecase/user_usecase.go

echo "Done formatting files"
