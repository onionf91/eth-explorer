package service

import (
	"github.com/gin-gonic/gin"
)

type ExplorerService struct {
}

func NewExplorerService() *ExplorerService {
	exp := &ExplorerService{}
	return exp
}

func (exp *ExplorerService) GetBlockListHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		limit := c.DefaultQuery("limit", "10")
		c.JSON(200, gin.H{
			"limit": limit,
		})
	}
}

func (exp *ExplorerService) GetBlockByIdHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		c.JSON(200, gin.H{
			"id": id,
		})
	}
}

func (exp *ExplorerService) GetTransactionByTxHashHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		txHash := c.Param("txHash")
		c.JSON(200, gin.H{
			"tx_hash": txHash,
		})
	}
}
