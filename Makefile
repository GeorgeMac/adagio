.PHONY: install
install: ## Install adagio and adagiod
	go install -mod=vendor ./...

.PHONY: test
test: ## Run test suite
	go test -cover -race -mod=vendor ./...

.PHONY: test-with-integrations
test-with-integrations: ## Run test suite with integrations (i.e. etcd)
	go test -cover -count 5 -race -mod=vendor -tags etcd ./...

.PHONY: deps
deps: ## Fetch and vendor dependencies
	go mod vendor

.PHONY: protobuf
protobuf: protobuf-deps ## Build protocol buffers into twirp model and service definitions
	protoc --go_out=paths=source_relative:. ./pkg/adagio/adagio.proto
	protoc -I. --twirp_out=. --go_out=. ./pkg/rpc/controlplane/service.proto

protobuf-deps: ## Fetch protobuf dependencies
	@go get github.com/twitchtv/twirp/protoc-gen-twirp
	@go get github.com/golang/protobuf/protoc-gen-go

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
