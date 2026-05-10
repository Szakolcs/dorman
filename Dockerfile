# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS base
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

FROM base AS dev
RUN go install github.com/air-verse/air@latest
EXPOSE 8080
CMD ["air", "-c", ".air.toml"]

FROM base AS builder
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bin/app ./cmd

FROM alpine:3.21 AS prod
WORKDIR /app
COPY --from=builder /app/bin/app /app/app
EXPOSE 8080
ENV PORT=8080
CMD ["/app/app"]
