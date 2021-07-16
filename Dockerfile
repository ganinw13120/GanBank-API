FROM golang:1.16-alpine

RUN apk add --no-cache git

WORKDIR /go/src/project/

COPY . /go/src/project/

RUN go build -v -o main.go