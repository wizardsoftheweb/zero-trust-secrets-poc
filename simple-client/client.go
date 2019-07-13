package main

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

const clientPort = 4747

type Config struct {
	secrets []string
}



func main() {
	cwd, _ := os.Getwd()
	EnsureKeyFilesExist(cwd)
	GlobalState := &Config{}
	r := gin.Default()
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})g
	_ = r.Run(fmt.Sprintf(":%d", clientPort))
}
