FROM golang:alpine

RUN apk update && apk add --no-cache git

WORKDIR /app

COPY go* ./

RUN go mod download

COPY . .
