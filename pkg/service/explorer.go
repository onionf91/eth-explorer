package service

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"log"
	"math/big"
	"os"
)

type BlockHeader struct {
	BlockNumber uint64 `json:"block_number"`
	BlockHash   string `json:"block_hash"`
	BlockTime   uint64 `json:"block_time"`
	ParentHash  string `json:"parent_hash"`
}

type Block struct {
	BlockHeader
	TxHashList []string `json:"transactions"`
}

type Transaction struct {
	Hash    string      `json:"tx_hash"`
	From    string      `json:"from"`
	To      string      `json:"to"`
	Nonce   uint64      `json:"nonce"`
	Data    string      `json:"data"`
	Value   string      `json:"value"`
	LogList []*EventLog `json:"logs"`
}

type EventLog struct {
	Index uint   `json:"index"`
	Data  string `json:"data"`
}

type ExplorerService struct {
	client *ethclient.Client
	rdb    *redis.Client
}

func NewExplorerService() *ExplorerService {
	client, err := ethclient.Dial(os.Getenv("ETH_EXPLORER_RPC_ENDPOINT"))
	if err != nil {
		log.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("ETH_EXPLORER_REDIS_ENDPOINT"),
		Password: os.Getenv("ETH_EXPLORER_REDIS_PASSWORD"),
		DB:       0, // use default DB
	})
	exp := &ExplorerService{
		client: client,
		rdb:    rdb,
	}
	return exp
}

func (exp *ExplorerService) GetBlockListHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		blockNumber := exp.queryLatestBlockNumber()
		if blockNumber == nil {
			c.JSON(500, gin.H{
				"reason": "query latest block number failed",
			})
			return
		}
		limit := new(big.Int)
		limit, ok := limit.SetString(c.DefaultQuery("limit", "10"), 10)
		if !ok {
			c.JSON(400, gin.H{
				"reason": "invalid limit parameter",
			})
			return
		}
		headerList := make([]*BlockHeader, 0)
		for index := big.NewInt(0); index.Cmp(limit) < 0; index.Add(index, big.NewInt(1)) {
			header := exp.queryBlockHeaderByNumber(blockNumber)
			if header == nil {
				c.JSON(500, gin.H{
					"reason": "query block header failed",
				})
				return
			}
			headerList = append(headerList, header)
			blockNumber.Sub(blockNumber, big.NewInt(1))
		}
		c.JSON(200, headerList)
	}
}

func (exp *ExplorerService) GetBlockByIdHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		blockNumber := new(big.Int)
		blockNumber, ok := blockNumber.SetString(c.Param("id"), 10)
		if !ok {
			c.JSON(400, gin.H{
				"reason": "invalid block id",
			})
			return
		}
		block := exp.queryBlockByNumber(blockNumber)
		if block == nil {
			c.JSON(500, gin.H{
				"reason": "query block failed",
			})
			return
		}
		c.JSON(200, block)
	}
}

func (exp *ExplorerService) GetTransactionByTxHashHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		tx := exp.queryTransactionByHash(common.HexToHash(c.Param("txHash")))
		if tx == nil {
			c.JSON(500, gin.H{
				"reason": "query transaction failed",
			})
			return
		}
		c.JSON(200, tx)
	}
}

func (exp *ExplorerService) queryDataFromCache(key string, dist any) bool {
	val, err := exp.rdb.Get(context.Background(), key).Result()
	if err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(val), dist); err != nil {
		log.Println("unmarshal json to struc failed", err)
		return false
	}
	log.Println("query data from redis")
	return true
}

func (exp *ExplorerService) cacheData(key string, data any) bool {
	val, err := json.Marshal(data)
	if err != nil {
		log.Println("marshal struc to json failed", err)
		return false
	}
	if err := exp.rdb.Set(context.Background(), key, val, 0).Err(); err != nil {
		log.Println("cache data to redis failed", err)
		return false
	}
	// log.Println("cache data to redis")
	return true
}

func (exp *ExplorerService) queryLatestBlockNumber() *big.Int {
	blockNumber, err := exp.client.BlockNumber(context.Background())
	if err != nil {
		log.Println("query latest block number failed", err)
		return nil
	}
	return new(big.Int).SetUint64(blockNumber)
}

func (exp *ExplorerService) queryBlockHeaderByNumber(number *big.Int) *BlockHeader {
	key := "header_" + number.String()
	entity := &BlockHeader{}
	if suc := exp.queryDataFromCache(key, entity); suc {
		return entity
	}
	header, err := exp.client.HeaderByNumber(context.Background(), number)
	if err != nil {
		log.Println("query block header failed", err)
		return nil
	}
	entity.BlockNumber = header.Number.Uint64()
	entity.BlockHash = header.Hash().String()
	entity.BlockTime = header.Time
	entity.ParentHash = header.ParentHash.String()
	exp.cacheData(key, entity)
	return entity
}

func (exp *ExplorerService) queryBlockByNumber(number *big.Int) *Block {
	key := "block_" + number.String()
	entity := &Block{}
	if suc := exp.queryDataFromCache(key, entity); suc {
		return entity
	}
	block, err := exp.client.BlockByNumber(context.Background(), number)
	if err != nil {
		log.Println("query block failed", err)
		return nil
	}
	txHashList := make([]string, 0)
	for _, tx := range block.Transactions() {
		txHashList = append(txHashList, tx.Hash().String())
	}
	entity.BlockNumber = block.Header().Number.Uint64()
	entity.BlockHash = block.Header().Hash().String()
	entity.BlockTime = block.Header().Time
	entity.ParentHash = block.Header().ParentHash.String()
	entity.TxHashList = txHashList
	exp.cacheData(key, entity)
	return entity
}

func (exp *ExplorerService) queryTransactionByHash(hash common.Hash) *Transaction {
	key := "tx_" + hash.String()
	entity := &Transaction{}
	if suc := exp.queryDataFromCache(key, entity); suc {
		return entity
	}
	tx, _, err := exp.client.TransactionByHash(context.Background(), hash)
	if err != nil {
		log.Println("query transaction failed", err)
		return nil
	}
	receipt, err := exp.client.TransactionReceipt(context.Background(), tx.Hash())
	if err != nil {
		log.Println("query transaction receipt failed", err)
	}
	chainId, err := exp.client.NetworkID(context.Background())
	if err != nil {
		log.Println("derive chain id failed", err)
		return nil
	}
	msg, err := tx.AsMessage(types.NewEIP155Signer(chainId), receipt.BlockNumber)
	if err != nil {
		log.Println("unpack message from transaction failed", err)
		return nil
	}
	to := ""
	if msg.To() != nil {
		to = msg.To().String()
	}
	logList := make([]*EventLog, 0)
	for _, eLog := range receipt.Logs {
		logList = append(logList, &EventLog{
			Index: eLog.Index,
			Data:  common.BytesToHash(eLog.Data).String(),
		})
	}
	entity.Hash = tx.Hash().String()
	entity.From = msg.From().String()
	entity.To = to
	entity.Nonce = tx.Nonce()
	entity.Data = common.BytesToHash(tx.Data()).String()
	entity.Value = tx.Value().String()
	entity.LogList = logList
	exp.cacheData(key, entity)
	return entity
}
