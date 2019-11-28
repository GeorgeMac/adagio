FROM golang:1.13-alpine AS base

RUN apk update && apk add make git

WORKDIR /workspace

ADD go.mod go.mod
ADD go.sum go.sum

RUN go mod download

ADD cmd/ cmd/
ADD pkg/ pkg/
ADD Makefile .

RUN make build

FROM alpine:3.10

COPY --from=base /workspace/bin/* /usr/bin/
