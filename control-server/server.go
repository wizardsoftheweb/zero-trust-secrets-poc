package main

import (
	"github.com/gin-gonic/gin"

	ginprometheus "github.com/zsais/go-gin-prometheus"
)

type RandoRequest struct {
	Count  int    `json: count`
	KvHost string `json: kv_host`
	Key    string `json: key`
}

func main() {
	r := gin.Default()
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.POST("/rando", func(c *gin.Context) {
		var request RandoRequest
		_ = c.BindJSON(&request)
		RandomStrings := make([]string, request.Count)
		for index := 0; index < request.Count; index++ {
			RandomStrings[index], _ = GenerateRandomString(47)
		}
		c.JSON(200, gin.H{
			"message": RandomStrings,
		})
	})
	_ = r.Run() // listen and serve on 0.0.0.0:8080
}
