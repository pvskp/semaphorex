FROM golang:1.22.3-alpine as builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -o vehicle-discovery ./cmd/vehicle-discovery/main.go

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/vehicle-discovery .
EXPOSE 8001
CMD ["./vehicle-discovery"]
