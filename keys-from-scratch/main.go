package main

import (
	"crypto"
	"log"
	"os"
	"time"

	"github.com/pkg/profile"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

const (
	name              = "CJ Harries"
	email             = "cj@wotw.pro"
	comment           = "Home Brew ZTS PoC"
	baseHash          = crypto.SHA512
	baseCipher        = packet.CipherCAST5
	commonCompression = packet.CompressionZLIB
	baseLevel         = packet.BestCompression
	minimumRsaBits    = 4096
)

func newGenericConfig() *packet.Config {
	return &packet.Config{
		DefaultHash:            baseHash,
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
	cwd, _ := os.Getwd()
	defer profile.Start(profile.ProfilePath(cwd)).Stop()
	_ = generateEntity()
}
