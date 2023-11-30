### Get the latest block height being processed
### GET /v3/info/lastnumber
```shell
curl --location --request POST 'http://127.0.0.1:8089/v3/info/lastnumber'
```

## DRC20
### Get all token information of DRC20
### GET /v3/drc20/all
```shell
curl --location --request GET 'curl --location 'http://127.0.0.1:8089/v3/drc20/all' \
--header 'Content-Type: application/json' \
--data '{"limit": 1500, "offset":0}'
```

### To obtain DRC20 token information, follow the tick
### GET /v3/drc20/tick
```shell
curl --location --request GET 'curl --location 'http://127.0.0.1:8089/v3/drc20/tick' \
--header 'Content-Type: application/json' \
--data '{"tick": "UNIX"}'
```

### Get DRC20 transaction information, according to block height
### GET /v3/drc20/order/number
```shell
curl --location --request GET 'curl --location 'http://127.0.0.1:8089/v3/drc20/order/number' \
--header 'Content-Type: application/json' \
--data '{"number": 100, "limit": 1500, "offset":0}'
```