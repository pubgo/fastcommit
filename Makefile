Project=fastcommit
Base=github.com/pubgo/funk
VERSION := $(shell git tag --sort=committerdate | tail -n 1)
GIT_COMMIT := $(shell git describe --always --abbrev=7 --dirty)
BUILD_TIME := $(shell date "+%F %T")
BRANCH_NAME=$(shell git rev-parse --abbrev-ref HEAD)

LDFLAGS=-ldflags " \
-X '${Base}/version.buildTime=${BUILD_TIME}' \
-X '${Base}/version.commitID=${GIT_COMMIT}' \
-X '${Base}/version.version=${VERSION}' \
-X '${Base}/version.project=${Project}' \
"

SHELL := /bin/bash

-include .env

.EXPORT_ALL_VARIABLES:

DEV_MSG ?= "dev"
TEST_MSG ?= "test"

.PHONY: build
build:
	go build ${LDFLAGS} -v -o bin/main main.go

vet:
	@go vet ./...
