package main

import (
	"log"
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
