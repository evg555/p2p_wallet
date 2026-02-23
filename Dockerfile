# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
ARG VERSION=dev
ARG COMMIT_SHA=unknown
ARG BUILD_TIME=unknown
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w -X main.VERSION=${VERSION} -X main.COMMIT_SHA=${COMMIT_SHA} -X main.BUILD_TIME=${BUILD_TIME}" \
    -o /out/auth ./cmd/auth

FROM alpine:3.22 AS runtime
WORKDIR /app

RUN adduser -D -u 10001 appuser

COPY --from=builder /out/auth .
COPY spec/openapi/users.yaml ./spec/openapi/users.yaml

USER appuser
EXPOSE 8080

CMD ["./auth"]
