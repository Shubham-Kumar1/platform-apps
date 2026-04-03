# ---- Build Stage ----
FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git ca-certificates
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" -o /event-service ./cmd/server

# ---- Runtime Stage ----
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /event-service /event-service
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/event-service"]
