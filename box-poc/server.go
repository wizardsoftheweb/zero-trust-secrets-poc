package main

import (
	"encoding/base64"
	"fmt"

	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

type OpenAndDecryptRequest struct {
	Message       string `json:"message"`
	EncryptPubKey string `json:"encrypt_pub"`
	SignPubKey    string `json:"sign_pub"`
	SignedMessage string `json:"signed_and_sealed"`
}

func BuildServer() {
	boxAndSign, _ := LoadOrCreateBoxAndSign(".")
	fmt.Println(boxAndSign.encryptionPubKey)
	r := gin.Default()
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/encrypt-key", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "success",
			"pub":     base64.StdEncoding.EncodeToString(boxAndSign.encryptionPubKey[:]),
		})
	})
	r.GET("/sign-key", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "success",
			"pub":     base64.StdEncoding.EncodeToString(boxAndSign.signingPubKey[:]),
		})
	})
	r.POST("/open-and-decrypt", func(c *gin.Context) {
		var request OpenAndDecryptRequest
		_ = c.BindJSON(&request)
		data, _ := base64.StdEncoding.DecodeString(request.EncryptPubKey)
		sendEncPubKey := new([32]byte)
		copy(sendEncPubKey[:], data)
		data, _ = base64.StdEncoding.DecodeString(request.SignPubKey)
		sendSigPubKey := new([32]byte)
		copy(sendSigPubKey[:], data)
		signedMessage, _ := base64.StdEncoding.DecodeString(request.SignedMessage)
		secret := boxAndSign.OpenAndDecrypt(signedMessage, sendSigPubKey, sendEncPubKey)
		c.JSON(200, gin.H{
			"message": secret,
		})
	})
	_ = r.Run() // listen and serve on 0.0.0.0:8080
}
