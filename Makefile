REGISTRY ?= quay.io
IMAGE_NAMESPACE ?= rhsyseng
IMAGE_NAME ?= cluster0-operators
IMAGE_URL ?= $(REGISTRY)/$(IMAGE_NAMESPACE)/$(IMAGE_NAME)
TAG ?= latest

.PHONY: build run get-dependencies build-image push-image build-and-push-image

build: get-dependencies
	$(info Building Linux, Mac and Windows binaries)
	mkdir -p ./out/
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-X 'github.com/rhsyseng/cluster0-operators/pkg/version.gitCommit=$(shell git rev-parse HEAD)' -X 'github.com/rhsyseng/cluster0-operators/pkg/version.buildTime=$(shell date +%Y-%m-%dT%H:%M:%SZ)'" -o ./out/cluster0-operators-linux-amd64 cmd/main.go
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-X 'github.com/rhsyseng/cluster0-operators/pkg/version.gitCommit=$(shell git rev-parse HEAD)' -X 'github.com/rhsyseng/cluster0-operators/pkg/version.buildTime=$(shell date +%Y-%m-%dT%H:%M:%SZ)'" -o ./out/cluster0-operators-linux-arm64 cmd/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'github.com/rhsyseng/cluster0-operators/pkg/version.gitCommit=$(shell git rev-parse HEAD)' -X 'github.com/rhsyseng/cluster0-operators/pkg/version.buildTime=$(shell date +%Y-%m-%dT%H:%M:%SZ)'" -o ./out/cluster0-operators-darwin-amd64 cmd/main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'github.com/rhsyseng/cluster0-operators/pkg/version.gitCommit=$(shell git rev-parse HEAD)' -X 'github.com/rhsyseng/cluster0-operators/pkg/version.buildTime=$(shell date +%Y-%m-%dT%H:%M:%SZ)'" -o ./out/cluster0-operators-darwin-arm64 cmd/main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-X 'github.com/rhsyseng/cluster0-operators/pkg/version.gitCommit=$(shell git rev-parse HEAD)' -X 'github.com/rhsyseng/cluster0-operators/pkg/version.buildTime=$(shell date +%Y-%m-%dT%H:%M:%SZ)'" -o ./out/cluster0-operators-windows-amd64.exe cmd/main.go
run: get-dependencies
	go run cmd/main.go
get-dependencies:
	$(info Downloading dependencies)
	go mod download
build-image:
	podman build . -f Dockerfile --build-arg gitCommit=$(shell git rev-parse HEAD) -t ${IMAGE_URL}:${TAG}
push-image:
	podman push ${IMAGE_URL}:${TAG}
build-and-push-image: build-image push-image