FROM golang:1.26-alpine AS builder

WORKDIR /build


COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /build/app ./src/cmd/main.go

FROM alpine:3.20

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /build/app .

EXPOSE 8080

CMD ["./app"]