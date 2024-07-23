FROM golang:1.22.5-alpine as buildbase

WORKDIR /go/src/github.com/Dmytro-Hladkykh/usdt-listener-svc

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /usr/local/bin/usdt-listener-svc .

FROM alpine:3.9

RUN apk add --no-cache ca-certificates

COPY --from=buildbase /usr/local/bin/usdt-listener-svc /usr/local/bin/usdt-listener-svc

ENTRYPOINT ["usdt-listener-svc"]
