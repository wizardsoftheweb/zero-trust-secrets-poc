package main

import (
	"github.com/gin-gonic/gin"

	ginprometheus "github.com/zsais/go-gin-prometheus"
)

func main() {
	r := gin.Default()
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	r.GET("/rando", func(c *gin.Context) {
		response, _ := GenerateRandomString(47)
		c.JSON(200, gin.H{
			"message": response,
		})
	})
	_ = r.Run() // listen and serve on 0.0.0.0:8080
}
