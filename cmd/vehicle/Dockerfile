FROM golang:1.22.3-alpine AS builder
WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN go build -mod=readonly -o vehicle ./cmd/vehicle/

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /app/vehicle .
EXPOSE 50050-50058
CMD ["./vehicle"]
