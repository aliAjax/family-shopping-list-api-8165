FROM golang:1.22-alpine AS builder

WORKDIR /src
COPY go.mod ./
COPY . .

RUN go mod tidy && CGO_ENABLED=0 GOOS=linux go build -trimpath -o /out/server ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata wget
WORKDIR /app
COPY --from=builder /out/server ./server
COPY migrations ./migrations

EXPOSE 8080

CMD ["./server"]
