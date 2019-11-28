gateway_path = "$$GOPATH/pkg/mod/$(shell go list -m github.com/grpc-ecosystem/grpc-gateway | head | sed 's/ /@/')"

.PHONY: install
install: ## Install adagio and adagiod
	go install ./...

.PHONY: build
build: ## Build adagio and adagiod into local bin dir
	@mkdir -p bin/
	go build -o bin/adagio  cmd/adagio/*.go
	go build -o bin/adagiod cmd/adagiod/*.go
	go build -o bin/adagiogw cmd/adagiogw/*.go

.PHONY: test
test: ## Run test suite
	go test -cover -race ./...

.PHONY: test-with-integrations
test-with-integrations: ## Run test suite with integrations (i.e. etcd)
	@hack/integration-test.sh

.PHONY: deps
deps: ## Fetch and vendor dependencies
	go mod download

.PHONY: fmt
fmt: ## Run go fmt -s all over the shop
	@gofmt -s -w $(shell find . -name "*.go")

.PHONY: protobuf
protobuf: protobuf-deps ## Build protocol buffers into model and grpc service definitions
	protoc --go_out=paths=source_relative:. ./pkg/adagio/adagio.proto
	protoc -I. -I$(gateway_path)/third_party/googleapis --go_out=plugins=grpc:. ./pkg/rpc/controlplane/service.proto
	protoc -I. -I$(gateway_path)/third_party/googleapis --grpc-gateway_out=logtostderr=true:. ./pkg/rpc/controlplane/service.proto
	protoc -I. -I$(gateway_path)/third_party/googleapis --swagger_out=logtostderr=true:. ./pkg/rpc/controlplane/service.proto

protobuf-deps: ## Fetch protobuf dependencies
	@go get github.com/golang/protobuf/{proto,protoc-gen-go}
	@go get google.golang.org/grpc
	@go get github.com/grpc-ecosystem/grpc-gateway

.PHONY: docker-build
docker-build: ## Build docker images
	docker build -t georgemac/adagio:`git describe --always --dirty` -f docker/Dockerfile .

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
