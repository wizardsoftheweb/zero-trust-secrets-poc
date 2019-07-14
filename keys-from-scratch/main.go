package main

import (
	"log"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

const (
	name              = "CJ Harries"
	email             = "cj@wotw.pro"
	comment           = "Home Brew ZTS PoC"
	minimumRsaBits    = 4096
	baseCipher        = packet.CipherAES256
	commonCompression = packet.CompressionZLIB
	baseLevel         = packet.BestSpeed
)

func newGenericConfig() *packet.Config {
	return &packet.Config{
		DefaultCipher:          baseCipher,
		DefaultCompressionAlgo: commonCompression,
		CompressionConfig: &packet.CompressionConfig{
			Level: baseLevel,
		},
		RSABits: minimumRsaBits,
	}
}

func generateEntity() error {
	start := time.Now()
	entity, err := openpgp.NewEntity(name, email, comment, newGenericConfig())
	if nil != err {
		return err
	}
	end := time.Now()
	log.Printf("duration: %v", end.Sub(start))
	log.Printf("%#v", entity)
	return nil
}

func main() {
	_ = generateEntity()
}
