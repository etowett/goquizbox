#Compile stage
FROM golang:1.19.3-alpine AS compiler

# Add required packages
RUN apk add  --no-cache --update git curl bash

ENV CGO_ENABLED 0 \
    GOOS=linux

WORKDIR /app

ADD go.mod go.sum ./
RUN go mod download

ADD . .

ENV CGO_ENABLED=0

ARG TAGS
ARG BUILD_ID
ARG BUILD_TAG
RUN go build -tags=${TAGS} -trimpath "-ldflags=-s -w -X=goquizbox/internal/buildinfo.BuildID=${BUILD_ID} -X=goquizbox/internal/buildinfo.BuildTag=${BUILD_TAG} -extldflags=-static" -o goquizbox cmd/server/main.go

# Run stage
FROM alpine:3.16
RUN apk update && \
    apk add mailcap tzdata && \
    rm /var/cache/apk/*
WORKDIR /app
COPY --from=compiler /app/goquizbox .
CMD ["/app/goquizbox"]
