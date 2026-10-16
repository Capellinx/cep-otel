FROM golang:1.23-alpine AS builder

ARG SERVICE

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/app ./cmd/${SERVICE}

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app

COPY --from=builder /app/bin/app /app/app

ENTRYPOINT ["/app/app"]
