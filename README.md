# Eth Explorer

## Configuration

```bash
export ETH_EXPLORER_RPC_ENDPOINT=https://...
export ETH_EXPLORER_REDIS_ENDPOINT=localhost:6379
export ETH_EXPLORER_REDIS_PASSWORD=
export ETH_EXPLORER_MYSQL_DNS=user:pass@tcp(127.0.0.1:3306)/dbname?charset\=utf8mb4&parseTime\=True&loc\=Local
```

## API Service

Execute api service:

```bash
go run cmd/api/main.go
```

### Get Newest Block Headers

This api will query the newest block headers and return
block information without transaction hash. The default
value of `limit` is `10`.

`GET /blocks?limit=n`

* response:

```json
[
  {
    "block_hash": "0x91fc067a98b09dca83cb5d669272fdd7601e4986a3b801aeb34cca842215d5c8",
    "block_number": 18093543,
    "block_time": 1648887630,
    "parent_hash": "0x4ad15bb01da56d7e8bede458a567b41fe9bfade240f6806f802e92b1446c70c6"
  },
  {
    "block_hash": "0x4ad15bb01da56d7e8bede458a567b41fe9bfade240f6806f802e92b1446c70c6",
    "block_number": 18093542,
    "block_time": 1648887627,
    "parent_hash": "0xfd49c2da21254b1fbdb86e109ac53fa9cf61f1fff996b6e728ec0aa94775f1e4"
  },
  {
    "block_hash": "0xfd49c2da21254b1fbdb86e109ac53fa9cf61f1fff996b6e728ec0aa94775f1e4",
    "block_number": 18093541,
    "block_time": 1648887624,
    "parent_hash": "0x8ed3d3b4667aceff6b121699dcf57359f3f600feb38f1b7f49e592e7fec25998"
  }
]
```

### Get Block by Block Number(ID)

This api will query a specific block by its block
number(ID). The response payload will include all hash
values of transactions that belong to this block.

`GET /blocks/:id`

* response:

```json
{
  "block_hash": "0xfd49c2da21254b1fbdb86e109ac53fa9cf61f1fff996b6e728ec0aa94775f1e4",
  "block_number": 18093541,
  "block_time": 1648887624,
  "parent_hash": "0x8ed3d3b4667aceff6b121699dcf57359f3f600feb38f1b7f49e592e7fec25998",
  "transactions": [
    "0x659197ea09f315964db1110e870a98ffcb88c0e97cf30b816fa6c3c9c626e124",
    "0x898e8f5bfae1ad5b76670e58b7e8c9f0c760976fdb5ecf517cf743688b7bd402",
    "0x0587f02cdc7cc5b7f8523ccb5cbe0fbf87d66f6d4ad364ed63d6f7c49f04ef8d",
    "0x4d657e23fd82f58e7aa62ef51b21ea1d2b78504ae75a20a9359c7b1f58aeb6e1",
    "0x24306f21fc13674385dba5337a4511d7076fbcd5697c484f2cc611a72f417eb7",
    "0x1e597118c34e132b6f55ae03f6900d404fc0ebc4a2d236a4eb23b3d5a4c5368c",
    "0xbd9f40a5a7b5b4887b92849253a7129753bdc7c03865e588e427abe4477f5adb",
    "0xf7a1834c1c427b6962d059b72e1f4af807e92c0f5292f5c230dfc9bd80b0ee9a",
    "0x8befb6ec1715512cdf791a838a8e88501d5b8853f4ea1ced82c9a118324d0ec2",
    "0xfde044779ab49260d23b712ab7f06aea2f0bb0171dacb5b9a2db4b010ce2772a",
    "0xd652ba131d20924ab64c1ac244b53b8cbd9fbccb58457eeefc8d5c764d45c1bc"
  ]
}
```

### Get Transaction by Hash

This api will query a specific transaction by its hash
value. The response payload will include event logs.

`GET /transaction/:txHash`

* response:

```json
{
  "data": "0x0000000000000000000000000000000000000000000000000022d1f523719000",
  "from": "0x084b2C96E961865F33Cdef47502affd5f5099e28",
  "logs": [
    {
      "data": "0x0000000000000000000000000000000000000000000000000022d1f523719000",
      "index": 9
    },
    {
      "data": "0x0000000000000000000000000000000000000000000000000022d1f523719000",
      "index": 9
    }
  ],
  "nonce": 3419,
  "to": "0x5527CFdc0FCd9F971af71599Fb636DE3EFbf1f87",
  "tx_hash": "0xbd9f40a5a7b5b4887b92849253a7129753bdc7c03865e588e427abe4477f5adb",
  "value": "0"
}
```

## Indexer Service

TBD...

## Enhancement

1. How to reduce invocation of RPC?
   * Using `Redis` to cached queried data.