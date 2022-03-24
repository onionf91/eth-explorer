package main

import (
	"github.com/gin-gonic/gin"
	"github.com/onionf91/eth-explorer/pkg/service"
)

func main() {

	router := gin.Default()
	explorer := service.NewExplorerService()

	router.GET("/ping", explorer.PingHandler())

	router.Run("localhost:8080")
}
