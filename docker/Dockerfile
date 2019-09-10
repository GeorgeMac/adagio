FROM golang:1.12-alpine AS base

RUN apk update && apk add make git

WORKDIR /workspace

ADD . /workspace

RUN make build

FROM alpine:3.10

COPY --from=base /workspace/bin/* /usr/bin/
