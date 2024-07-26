FROM golang:1.22.5-alpine as buildbase

WORKDIR /go/src/github.com/Dmytro-Hladkykh/usdt-listener-svc

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -o /usr/local/bin/usdt-listener-svc .

FROM alpine:latest

RUN apk add --no-cache ca-certificates

COPY --from=buildbase /usr/local/bin/usdt-listener-svc /usr/local/bin/usdt-listener-svc

ENTRYPOINT ["/usr/local/bin/usdt-listener-svc"]
