FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/taskforge ./cmd/taskforge
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/taskforge-worker ./cmd/taskforge-worker

FROM alpine:3.20
RUN apk add --no-cache ca-certificates bash curl
WORKDIR /app
COPY --from=builder /bin/taskforge /app/taskforge
COPY --from=builder /bin/taskforge-worker /app/taskforge-worker
COPY migrations /app/migrations
COPY config /app/config
EXPOSE 8080
ENTRYPOINT ["/app/taskforge"]
