FROM golang:latest

RUN mkdir /app
ADD . /app
WORKDIR /app

RUN go build -o bin/main

CMD ["./bin/main"]