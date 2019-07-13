package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"

	"github.com/gin-gonic/gin"
	ginprometheus "github.com/zsais/go-gin-prometheus"
)

const (
	clientPort       = 4747
	secretCount      = 10
	secretsKey       = "/simple-client/secrets.json"
	controlServerUrl = "http://localhost:8080/rando"
)

var etcdHosts = []string{
	"http://127.0.0.1:2379/",
}

type Config struct {
	secrets []string
}

type RandoRequest struct {
	Count   int      `json:"count"`
	KvHosts []string `json:"kv_hosts"`
	KvKey   string   `json:"kv_key"`
	PubKey  string   `json:"pub_key"`
}

type RandoResponse struct {
	Secrets []string `json:"message"`
}

func loadPubKey(directory string) string {
	pubKeyFileName := GetKeyFileName(directory, GpgKeyTypePub)
	rawContents, _ := ioutil.ReadFile(pubKeyFileName)
	contents := string(rawContents)
	contents = strings.TrimSpace(contents)
	contents = strings.ReplaceAll(contents, "\n", "\\n")
	return contents
}

func GenerateSecrets(directory string) []string {
	requestBody := &RandoRequest{
		Count:   secretCount,
		KvHosts: etcdHosts,
		KvKey:   secretsKey,
		PubKey:  loadPubKey(directory),
	}
	log.Printf("%#v", requestBody)
	buffer := new(bytes.Buffer)
	_ = json.NewEncoder(buffer).Encode(requestBody)
	request, _ := http.NewRequest("POST", controlServerUrl, buffer)
	client := &http.Client{}
	response, err := client.Do(request)
	if nil != err {
		log.Println("Unable to post request")
		log.Fatal(err)
	}
	responseBody, err := ioutil.ReadAll(response.Body)
	if nil != err {
		log.Fatal(err)
	}
	var parsedResponse RandoResponse
	err = json.Unmarshal(responseBody, &parsedResponse)
	if nil != err {
		log.Fatal(err)
	}
	return parsedResponse.Secrets
}

func BootstrapViper(directory string) {
	err := viper.AddSecureRemoteProvider(
		"etcd",
		etcdHosts[0],
		secretsKey,
		GetKeyFileName(directory, GpgKeyTypeSecret),
	)
	if nil != err {
		log.Fatal(err)
	}
	viper.SetConfigType("json")
	err = viper.ReadRemoteConfig()
	if nil != err {
		log.Fatal(err)
	}
	secrets := viper.GetStringSlice("secrets")
	if 0 == len(secrets) {
		log.Println("Generating secrets")
		GenerateSecrets(directory)
	}
}

func main() {
	cwd, _ := os.Getwd()
	EnsureKeyFilesExist(cwd)
	BootstrapViper(cwd)
	GlobalState := &Config{
		secrets: viper.GetStringSlice("secrets"),
	}
	r := gin.Default()
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"secrets": GlobalState.secrets,
		})
	})
	r.GET("/force-update", func(c *gin.Context) {
		GenerateSecrets(cwd)
		c.JSON(200, gin.H{
			"message": "Secrets were regenerated",
		})
	})
	_ = r.Run(fmt.Sprintf(":%d", clientPort))
}
