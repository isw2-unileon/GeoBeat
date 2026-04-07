FROM golang:1.24-alpine AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY backend/ ./backend/

RUN go build -o main ./backend/cmd/server

FROM alpine:latest

WORKDIR /root

COPY --from=build /app/main .

EXPOSE 8080

CMD ["./main"]