# Stage 1
FROM golang:1.26.0 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server

#Stage 2
FROM alpine:3.21.3
RUN apk add --no-cache curl
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
WORKDIR /app
COPY --from=builder /app/server .
COPY migrations ./migrations
RUN chown -R appuser:appgroup /app
USER appuser
EXPOSE 8080
CMD ["./server"]