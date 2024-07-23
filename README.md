# usdt-listener-svc

## Description

This service listens events of transfers USDT token in Ethereum, saves them in db and returns using API.

## Install

```
git clone https://github.com/Dmytro-Hladkykh/usdt-listener-svc
cd usdt-listener-svc
go build main.go
export KV_VIPER_FILE=./config.yaml
./main migrate up
./main run service
```

## Infura API key

1. Sign up to Infura
2. Configure your API key
3. Create .env in project root with API key

```
INFURA_API_KEY=your_api_key
```

## Running from docker

Make sure that docker installed.

Configure `docker-compose.yml` with entrypoint commands:

```
entrypoint:
      [
        "/bin/sh",
        "-c",
        "/usr/local/bin/usdt-listener-svc migrate up && /usr/local/bin/usdt-listener-svc run service",
      ]
```

Possible commands are:

```
/usr/local/bin/usdt-listener-svc migrate up
```

```
/usr/local/bin/usdt-listener-svc run service
```

```
/usr/local/bin/usdt-listener-svc migrate down
```

To run the service with Docker use:

```
docker-compose up --build
```

## Testing

To test usage you can use Postman.

### Create a GET with:

With this GET you can configure `number of page` with `page=` and `transactions per page` with `per_page=`:

```
http://localhost:80/usdt-listener-svc?page=1&per_page=10
```

In response you will get a 1st page with 10 USDT transactions :

```
{
        "ID": 1,
        "FromAddress": "0xbAd9ADa0E9633ED27Fa183dBdEceef76712a6Ee1",
        "ToAddress": "0x91C986709Bb4fE0763edF8E2690EE9d5019Bea4a",
        "Amount": "31232094295",
        "TransactionHash": "0xccd3a013c32d19a45b2e1623a46c29340ff867f56a7d7ed263553abe0260d9af",
        "BlockNumber": 20369176,
        "Timestamp": "2024-07-23T12:17:29.298918Z"
    },
    {
        "ID": 2,
        "FromAddress": "0x91C986709Bb4fE0763edF8E2690EE9d5019Bea4a",
        "ToAddress": "0x0cA50FD1F9f2c4DAEB8EC483241118c35e0e472d",
        "Amount": "31150890850",
        "TransactionHash": "0xccd3a013c32d19a45b2e1623a46c29340ff867f56a7d7ed263553abe0260d9af",
        "BlockNumber": 20369176,
        "Timestamp": "2024-07-23T12:17:29.301972Z"
    },
    ...
```

### Create a GET with:

```
http://localhost:80/usdt-listener-svc/1
```

In response you will get:

```
{
    "ID": 1,
    "FromAddress": "0xbAd9ADa0E9633ED27Fa183dBdEceef76712a6Ee1",
    "ToAddress": "0x91C986709Bb4fE0763edF8E2690EE9d5019Bea4a",
    "Amount": "31232094295",
    "TransactionHash": "0xccd3a013c32d19a45b2e1623a46c29340ff867f56a7d7ed263553abe0260d9af",
    "BlockNumber": 20369176,
    "Timestamp": "2024-07-23T12:17:29.298918Z"
}
```

## Running from Source

- Set up environment value with config file path `KV_VIPER_FILE=./config.yaml`
- Provide valid config file
- Launch the service with `migrate up` command to create database schema
- Launch the service with `run service` command

### Database

For services, we do use **_PostgresSQL_** database.
You can [install it locally](https://www.postgresql.org/download/) or use [docker image](https://hub.docker.com/_/postgres/).

### Third-party services

## Contact

Responsible Dmytro Hladkykh
The primary contact for this project is https://t.me/Dimo4kaaaaaa
