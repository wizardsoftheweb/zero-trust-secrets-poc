package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/prometheus/common/log"

	"github.com/gin-gonic/gin"

	"github.com/xordataexchange/crypt/config"

	ginprometheus "github.com/zsais/go-gin-prometheus"
)

type RandoRequest struct {
	Count   int      `json:"count"`
	KvHosts []string `json:"kv_hosts"`
	KvKey   string   `json:"kv_key"`
	PubKey  string   `json:"pub_key"`
}

func WriteGpgKeyToFile(gpgKey string) *os.File {
	file, err := ioutil.TempFile("", "")
	if nil != err {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(
		file.Name(),
		[]byte(strings.ReplaceAll(gpgKey, "\\n", "\n")),
		0666,
	)
	if nil != err {
		log.Fatal(err)
	}
	fmt.Println(file.Name())
	return file
}

func WriteValue(addresses []string, pubKey string, key string, value []string) {
	pubKeyFile := WriteGpgKeyToFile(pubKey)
	defer os.Remove(pubKeyFile.Name())
	fileReader, err := os.Open(pubKeyFile.Name())
	if nil != err {
		log.Fatal(err)
	}
	defer fileReader.Close()
	fmt.Println(addresses)
	configManager, err := config.NewEtcdConfigManager(addresses, fileReader)
	if nil != err {
		log.Fatal(err)
	}
	secretsJson := fmt.Sprintf(`{"secrets":["%s"]}`, strings.Join(value, `","`))
	fmt.Println(secretsJson)
	err = configManager.Set(key, []byte(secretsJson))
	if nil != err {
		log.Fatal(err)
	}
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
		fmt.Println(request)
		randomStrings := make([]string, request.Count)
		for index := 0; index < request.Count; index++ {
			randomStrings[index], _ = GenerateRandomString(47)
		}
		WriteValue(request.KvHosts, request.PubKey, request.KvKey, randomStrings)
		c.JSON(200, gin.H{
			"message": randomStrings,
		})
	})
	_ = r.Run() // listen and serve on 0.0.0.0:8080
}
