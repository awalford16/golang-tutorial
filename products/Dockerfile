FROM golang:latest

ENV GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy code files into container
COPY . .

RUN go build -o products .

WORKDIR /app

RUN cp /build/products .

EXPOSE 9090
CMD ["/app/products"]