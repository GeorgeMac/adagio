.PHONY: proto
proto: proto-deps
	protoc --twirp_out=. --go_out=. ./pkg/rpc/controlplane/service.proto

proto-deps:
	go get github.com/twitchtv/twirp/protoc-gen-twirp
	go get github.com/golang/protobuf/protoc-gen-go
