FROM golang:1.24-alpine AS deps
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM golang:1.24-alpine AS builder
COPY --from=deps /go/pkg/mod /go/pkg/mod
WORKDIR /app
COPY . .
RUN go build -o srd main.go

FROM scratch
COPY --from=builder /app/srd .
EXPOSE 8080
ENTRYPOINT ["./srd", "serve"]
