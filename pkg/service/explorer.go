package service

import (
	"github.com/gin-gonic/gin"
)

type ExplorerService struct {
	counter int
}

func NewExplorerService() *ExplorerService {
	exp := &ExplorerService{counter: 0}
	return exp
}

func (exp *ExplorerService) PingHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		exp.counter++
		c.JSON(200, gin.H{
			"message": "pong",
			"count":   exp.counter,
		})
	}
}
