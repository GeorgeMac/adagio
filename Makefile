.PHONY: install
install: ## Install adagio and adagiod
	go install ./...

.PHONY: protobuf
protobuf: protobuf-deps ## Build protocol buffers into twirp model and service definitions
	protoc --twirp_out=. --go_out=. ./pkg/rpc/controlplane/service.proto

protobuf-deps:
	go get github.com/twitchtv/twirp/protoc-gen-twirp
	go get github.com/golang/protobuf/protoc-gen-go

# Absolutely awesome: http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help
