package service

import (
	"context"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"log"
	"math/big"
	"os"
)

type ExplorerService struct {
	client *ethclient.Client
}

type Transaction struct {
	raw     *types.Transaction
	msg     *types.Message
	receipt *types.Receipt
}

func NewExplorerService() *ExplorerService {
	client, err := ethclient.Dial(os.Getenv("ETH_EXPLORER_RPC_ENDPOINT"))
	if err != nil {
		log.Fatal(err)
	}
	exp := &ExplorerService{client: client}
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
		blockList := make([]interface{}, limit.Int64())
		for index := big.NewInt(0); index.Cmp(limit) < 0; index.Add(index, big.NewInt(1)) {
			header := exp.queryBlockHeaderByNumber(blockNumber)
			blockNumber.Sub(blockNumber, big.NewInt(1))
			if header == nil {
				c.JSON(500, gin.H{
					"reason": "query block header failed",
				})
				return
			}
			blockList[index.Int64()] = gin.H{
				"block_number": header.Number.Uint64(),
				"block_hash":   header.Hash().String(),
				"block_time":   header.Time,
				"parent_hash":  header.ParentHash.String(),
			}
		}
		c.JSON(200, blockList)
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
		header := block.Header()
		block.Transactions().Len()
		txHashList := make([]interface{}, block.Transactions().Len())
		for index := 0; index < block.Transactions().Len(); index++ {
			txHashList[index] = block.Transactions()[index].Hash()
		}
		c.JSON(200, gin.H{
			"block_number": header.Number.Uint64(),
			"block_hash":   header.Hash().String(),
			"block_time":   header.Time,
			"parent_hash":  header.ParentHash.String(),
			"transactions": txHashList,
		})
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
		logList := make([]interface{}, len(tx.receipt.Logs))
		for index := 0; index < len(tx.receipt.Logs); index++ {
			logList[index] = gin.H{
				"index": tx.receipt.Logs[0].Index,
				"data":  common.BytesToHash(tx.receipt.Logs[0].Data),
			}
		}
		c.JSON(200, gin.H{
			"tx_hash": tx.raw.Hash().String(),
			"from":    tx.msg.From().String(),
			"to":      tx.msg.To().String(),
			"nonce":   tx.raw.Nonce(),
			"data":    common.BytesToHash(tx.raw.Data()),
			"value":   tx.raw.Value().String(),
			"logs":    logList,
		})
	}
}

func (exp *ExplorerService) queryLatestBlockNumber() *big.Int {
	blockNumber, err := exp.client.BlockNumber(context.Background())
	if err != nil {
		log.Println("query latest block number failed", err)
		return nil
	}
	return new(big.Int).SetUint64(blockNumber)
}

func (exp *ExplorerService) queryBlockHeaderByNumber(number *big.Int) *types.Header {
	header, err := exp.client.HeaderByNumber(context.Background(), number)
	if err != nil {
		log.Println("query block header failed", err)
		return nil
	}
	return header
}

func (exp *ExplorerService) queryBlockByNumber(number *big.Int) *types.Block {
	block, err := exp.client.BlockByNumber(context.Background(), number)
	if err != nil {
		log.Println("query block failed", err)
		return nil
	}
	return block
}

func (exp *ExplorerService) queryTransactionByHash(hash common.Hash) *Transaction {
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
	return &Transaction{
		raw:     tx,
		msg:     &msg,
		receipt: receipt,
	}
}
