FROM golang:1.20-alpine as buildbase

RUN apk add git build-base

WORKDIR /go/src/https://github.com/Dmytro-Hladkykh/usdt-listener-svc
COPY vendor .
COPY . .

RUN GOOS=linux go build  -o /usr/local/bin/usdt-listener-svc /go/src/https://github.com/Dmytro-Hladkykh/usdt-listener-svc


FROM alpine:3.9

COPY --from=buildbase /usr/local/bin/usdt-listener-svc /usr/local/bin/usdt-listener-svc
RUN apk add --no-cache ca-certificates

ENTRYPOINT ["usdt-listener-svc"]
