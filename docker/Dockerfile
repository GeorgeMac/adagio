FROM golang:1.12-alpine AS base

RUN apk update && apk add make git

WORKDIR /workspace

ADD go.mod go.mod
ADD go.sum go.sum

RUN go mod download

ADD . /workspace

RUN make build

FROM alpine:3.10

COPY --from=base /workspace/bin/* /usr/bin/

COPY pkg/rpc/controlplane/service.swagger.json /service.swagger.json
