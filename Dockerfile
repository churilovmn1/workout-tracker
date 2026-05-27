FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/bin/workout-tracker ./cmd/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/workout-tracker .
COPY migrations/ ./migrations/
COPY web/ ./web/

EXPOSE 8080
CMD ["./workout-tracker"]
