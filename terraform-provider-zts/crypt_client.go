package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/xordataexchange/crypt/config"
)

func newConfigManager(hosts []string, pubKeyPath string) *config.ConfigManager {
	fileReader, err := os.Open(pubKeyPath)
	if nil != err {
		log.Fatal(err)
	}
	defer fileReader.Close()
	configManager, err := config.NewEtcdConfigManager(
		hosts,
		fileReader,
	)
	if nil != err {
		log.Fatal(err)
	}
	return &configManager
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

func GenerateSecrets(controlServerUrl string, requestBody *RandoRequest) []string {
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
