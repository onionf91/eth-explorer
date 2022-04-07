package service

import (
	"context"
	"encoding/json"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/onionf91/eth-explorer/pkg/entity"
	"github.com/onionf91/eth-explorer/pkg/model"
	"github.com/reactivex/rxgo/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"math/big"
	"os"
	"sync"
)

type ExplorerService struct {
	client        *ethclient.Client
	rdb           *redis.Client
	gdb           *gorm.DB
	mutex         sync.Mutex
	associations  []*entity.BlockAssociation
	blockOf       map[string]*model.Block
	transactionOf map[string]*model.Transaction
	lastBlock     *model.Block
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
	gdb, err := gorm.Open(mysql.Open(os.Getenv("ETH_EXPLORER_MYSQL_DNS")), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	exp := &ExplorerService{
		client:        client,
		rdb:           rdb,
		gdb:           gdb,
		associations:  make([]*entity.BlockAssociation, 0),
		blockOf:       make(map[string]*model.Block),
		transactionOf: make(map[string]*model.Transaction),
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
		headerList := make([]*entity.BlockHeader, 0)
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

func (exp *ExplorerService) AutoMigrateSchema() {
	if err := exp.gdb.AutoMigrate(&model.Block{}, &model.Transaction{}); err != nil {
		log.Fatal(err)
	}
}

func (exp *ExplorerService) ScanBlockFrom(from uint64, parallels int) {
	to := exp.queryLatestBlockNumber()
	if from == 0 {
		log.Printf("latest block number is %v", to.Uint64())
		log.Println("please execute command with -start=block-number flag to specify which block starting scan from.")
		return
	}
	if from >= to.Uint64() {
		log.Printf("latest block number is %v", to.Uint64())
		log.Println("start block number must lower than latest one.")
		return
	}
	if result := exp.gdb.Where("number = ?", to.Uint64()).Find(&model.Block{}); result.RowsAffected != 0 {
		log.Printf("latest block %v already scanned.", to)
		return
	}
	<-exp.doScanBlockFrom(from, to.Uint64(), parallels).Run()
	exp.associateBlocksAndTransactions()
	log.Printf("total %v blocks scanned. persisting...", len(exp.associations))
	exp.gdb.Create(exp.lastBlock)
	log.Println("persist done.")
}

func (exp *ExplorerService) associateBlocksAndTransactions() {
	for _, association := range exp.associations {
		target := exp.blockOf[association.BlockHash]
		if parent, exists := exp.blockOf[association.ParentHash]; exists {
			target.Parent = parent
		}
		for _, uncleHash := range association.UncleHashList {
			if uncle, exists := exp.blockOf[uncleHash]; exists {
				target.Uncles = append(target.Uncles, *uncle)
			}
		}
		for _, txHash := range association.TxHashList {
			if tx, exists := exp.transactionOf[txHash]; exists {
				target.Transactions = append(target.Transactions, *tx)
			}
		}
	}
}

func (exp *ExplorerService) doScanBlockFrom(from uint64, to uint64, parallels int) rxgo.Observable {
	var withPool rxgo.Option
	if parallels != 0 {
		withPool = rxgo.WithPool(parallels)
	} else {
		withPool = rxgo.WithCPUPool()
	}
	return rxgo.Range(0, int(to-from)+1).Map(func(_ context.Context, v interface{}) (interface{}, error) {
		blockNumber := from + uint64(v.(int))
		blockModel, blockAssociation := exp.findBockModelByNumber(blockNumber)
		transactionModeOf := make(map[string]*model.Transaction)
		if blockAssociation != nil {
			for _, txHash := range blockAssociation.TxHashList {
				transactionModeOf[txHash] = exp.findTransactionModelByHash(txHash)
			}
		}
		exp.mutex.Lock()
		defer exp.mutex.Unlock()
		for txHash, txModel := range transactionModeOf {
			exp.transactionOf[txHash] = txModel
		}
		exp.blockOf[blockModel.Hash] = blockModel
		if blockAssociation != nil {
			exp.associations = append(exp.associations, blockAssociation)
		}
		if blockNumber == to {
			exp.lastBlock = blockModel
		}
		return v, nil
	}, withPool)
}

func (exp *ExplorerService) findTransactionModelByHash(hash string) *model.Transaction {
	transactionModel := model.Transaction{}
	if result := exp.gdb.Where("hash = ?", hash).Find(&transactionModel); result.RowsAffected != 0 {
		return &transactionModel
	}
	// TODO: scan data by rpc ?
	transactionModel.Hash = hash
	return &transactionModel
}

func (exp *ExplorerService) findBockModelByNumber(number uint64) (*model.Block, *entity.BlockAssociation) {
	blockModel := model.Block{}
	if result := exp.gdb.Where("number = ?", number).Find(&blockModel).RowsAffected; result != 0 {
		return &blockModel, nil
	}
	log.Printf("scan block : %v", number)
	block := exp.doQueryBlockByNumber(new(big.Int).SetUint64(number))
	if block == nil {
		log.Fatal("query block via grpc failed.")
	}
	txHashList := make([]string, 0)
	for _, tx := range block.Transactions() {
		txHashList = append(txHashList, tx.Hash().String())
	}
	uncleHashList := make([]string, 0)
	for _, uncle := range block.Uncles() {
		uncleHashList = append(uncleHashList, uncle.Hash().String())
	}
	blockModel.Difficulty = block.Difficulty().Uint64()
	blockModel.ExtraData = common.BytesToHash(block.Extra()).String()
	blockModel.GasLimit = block.GasLimit()
	blockModel.GasUsed = block.GasUsed()
	blockModel.Hash = block.Hash().String()
	blockModel.LogsBloom = common.BytesToHash(block.Bloom().Bytes()).String()
	blockModel.Miner = block.Coinbase().String()
	blockModel.MixHash = block.MixDigest().String()
	blockModel.Nonce = block.Nonce()
	blockModel.Number = number
	blockModel.ReceiptsRoot = block.ReceiptHash().String()
	blockModel.Sha3Uncles = block.UncleHash().String()
	blockModel.Size = float64(block.Size())
	blockModel.StateRoot = block.Root().String()
	blockModel.Timestamp = block.Time()
	return &blockModel, &entity.BlockAssociation{
		Block: entity.Block{
			BlockHeader: entity.BlockHeader{
				BlockHash:  block.Hash().String(),
				ParentHash: block.ParentHash().String(),
			},
			TxHashList: txHashList,
		},
		UncleHashList: uncleHashList,
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

func (exp *ExplorerService) queryBlockHeaderByNumber(number *big.Int) *entity.BlockHeader {
	key := "header_" + number.String()
	headerEntity := &entity.BlockHeader{}
	if suc := exp.queryDataFromCache(key, headerEntity); suc {
		return headerEntity
	}
	header, err := exp.client.HeaderByNumber(context.Background(), number)
	if err != nil {
		log.Println("query block header failed", err)
		return nil
	}
	headerEntity.BlockNumber = header.Number.Uint64()
	headerEntity.BlockHash = header.Hash().String()
	headerEntity.BlockTime = header.Time
	headerEntity.ParentHash = header.ParentHash.String()
	exp.cacheData(key, headerEntity)
	return headerEntity
}

func (exp *ExplorerService) queryBlockByNumber(number *big.Int) *entity.Block {
	key := "block_" + number.String()
	blockEntity := &entity.Block{}
	if suc := exp.queryDataFromCache(key, blockEntity); suc {
		return blockEntity
	}
	if block := exp.doQueryBlockByNumber(number); block != nil {
		txHashList := make([]string, 0)
		for _, tx := range block.Transactions() {
			txHashList = append(txHashList, tx.Hash().String())
		}
		blockEntity.BlockNumber = block.Header().Number.Uint64()
		blockEntity.BlockHash = block.Header().Hash().String()
		blockEntity.BlockTime = block.Header().Time
		blockEntity.ParentHash = block.Header().ParentHash.String()
		blockEntity.TxHashList = txHashList
		exp.cacheData(key, blockEntity)
		return blockEntity
	}
	return nil
}

func (exp *ExplorerService) doQueryBlockByNumber(number *big.Int) *types.Block {
	block, err := exp.client.BlockByNumber(context.Background(), number)
	if err != nil {
		log.Println("query block failed", err)
		return nil
	}
	return block
}

func (exp *ExplorerService) queryTransactionByHash(hash common.Hash) *entity.Transaction {
	key := "tx_" + hash.String()
	transactionEntity := &entity.Transaction{}
	if suc := exp.queryDataFromCache(key, transactionEntity); suc {
		return transactionEntity
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
	logEntityList := make([]*entity.EventLog, 0)
	for _, eLog := range receipt.Logs {
		logEntityList = append(logEntityList, &entity.EventLog{
			Index: eLog.Index,
			Data:  common.BytesToHash(eLog.Data).String(),
		})
	}
	transactionEntity.Hash = tx.Hash().String()
	transactionEntity.From = msg.From().String()
	transactionEntity.To = to
	transactionEntity.Nonce = tx.Nonce()
	transactionEntity.Data = common.BytesToHash(tx.Data()).String()
	transactionEntity.Value = tx.Value().String()
	transactionEntity.LogList = logEntityList
	exp.cacheData(key, transactionEntity)
	return transactionEntity
}
