package main

import (
	"github.com/gin-gonic/gin"
	"github.com/onionf91/eth-explorer/pkg/service"
)

func main() {

	router := gin.Default()
	explorer := service.NewExplorerService()

	router.GET("/blocks", explorer.GetBlockListHandler())
	router.GET("/blocks/:id", explorer.GetBlockByIdHandler())
	router.GET("/transaction/:txHash", explorer.GetTransactionByTxHashHandler())

	router.Run("localhost:8080")
}
