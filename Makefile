# Determine this makefile's path.
# Be sure to place this BEFORE `include` directives, if any.
THIS_FILE := $(lastword $(MAKEFILE_LIST))
THIS_DIR := $(dir $(realpath $(firstword $(MAKEFILE_LIST))))

TMP_DIR := $(shell mktemp -d)
REPO_PATH := github.com/hashicorp/dbw

tools:
	go generate -tags tools tools/tools.go

fmt:
	gofumpt -w $$(find . -name '*.go' | grep -v pb.go)

test: 
	go test -race -count=1 ./...

test-all: test-sqlite test-postgres

test-sqlite:
	 DB_DIALECT=sqlite go test -race -count=1 ./...

test-postgres:
	 DB_DIALECT=postgres DB_DSN="postgresql://go_db:go_db@localhost:9920/go_db?sslmode=disable"  go test -race -count=1 ./...

### db tags requires protoc-gen-go v1.20.0 or later
# GO111MODULE=on go get -u github.com/golang/protobuf/protoc-gen-go@v1.40
proto: protolint protobuild

protobuild:
	protoc \
	./internal/proto/local/dbtest/storage/v1/dbtest.proto \
	--proto_path=internal/proto/local \
	--go_out=:.
	
	@protoc-go-inject-tag -input=./internal/dbtest/dbtest.pb.go

protolint:
	@buf lint
	# if/when this becomes a public repo, we can add this check
	# @buf check breaking --against 'https://github.com/hashicorp/go-dbw.git#branch=main'