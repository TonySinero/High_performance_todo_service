FROM golang:alpine AS builder

COPY . /newFeatures/todo_service/
WORKDIR /newFeatures/todo_service/

RUN go mod download
RUN GOOS=linux go build -o ./.bin/service ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=0 /newFeatures/todo_service/.bin/service .

EXPOSE 8080

CMD ["./service"]