FROM golang:alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY . ./

RUN go mod download

COPY . .
