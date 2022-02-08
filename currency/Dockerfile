FROM golang:latest

LABEL maintainer="Adam <adam.walford@citrix.com>"

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ENV PORT=9090

RUN go build

CMD ["./currency"]