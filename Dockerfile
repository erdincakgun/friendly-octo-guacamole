FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o mock-service .

FROM scratch
COPY --from=builder /app/mock-service /mock-service
EXPOSE 8080
ENTRYPOINT ["/mock-service"]
