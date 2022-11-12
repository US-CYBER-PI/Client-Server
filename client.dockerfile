FROM golang:1.19-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY . ./

RUN go build -o /app/client-server .

FROM alpine

WORKDIR /app

COPY --from=builder /app/client-server ./client-server

CMD ["./client-server"]