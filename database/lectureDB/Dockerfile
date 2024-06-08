# SPDX-License-Identifier: Apache-2.0

### Builder

FROM golang:1.22-alpine3.20 as builder

RUN apk --no-cache update
RUN apk add --no-cache git clang llvm make gcc protobuf make
RUN apk add --update alpine-sdk

RUN mkdir /app
RUN mkdir /protobuf

WORKDIR /protobuf
COPY /protobuf .

WORKDIR /app
COPY /database/lectureDB .

RUN go install github.com/golang/protobuf/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN export CGO_ENABLED=1; export CC=gcc;

RUN go build -o lectureDB

### Make executable image

FROM alpine:3.18 as lectureDB

RUN echo "@community http://dl-cdn.alpinelinux.org/alpine/edge/community" | tee -a /etc/apk/repositories

RUN apk --no-cache update
RUN apk add bash

COPY --from=builder /app/lectureDB .
COPY --from=builder /app/courses.json .

CMD ["./lectureDB"]