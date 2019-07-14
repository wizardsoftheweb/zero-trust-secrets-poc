package main

import (
	"fmt"
	"log"
	"strings"
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

var (
	possibleCiphersForBenchmarks = []string{
		"packet.CipherAES128",
		"packet.CipherAES192",
		"packet.CipherAES256",
	}
	possibleCompressionAlgosForBenchmarks = []string{
		"packet.CompressionZLIB",
		"packet.CompressionZIP",
		"packet.CompressionNone",
	}
	possibleCompressionLevelsForBenchmarks = []string{
		"packet.BestSpeed",
		"packet.BestCompression",
		"packet.DefaultCompression",
		"packet.NoCompression",
	}
	possibleRsaBitsForBenchmarks = []string{"2048", "4096"}
)

func draftBenchmarks() {
	for _, cipher := range possibleCiphersForBenchmarks {
		for _, compressionAlgo := range possibleCompressionAlgosForBenchmarks {
			for _, level := range possibleCompressionLevelsForBenchmarks {
				for _, rsa := range possibleRsaBitsForBenchmarks {
					cipherName := strings.TrimPrefix(cipher, "packet.Cipher")
					compressionAlgoName := strings.TrimPrefix(compressionAlgo, "packet.Compression")
					levelName := strings.TrimPrefix(level, "packet.")
					fmt.Printf(
						`
func (s *CreationSuite) BenchmarkCreationOf%sx%sx%sx%s(c *C) {
	benchmarkSingleConfig(%s, %s, %s, %s, c)
}
`,
						cipherName,
						compressionAlgoName,
						levelName,
						rsa,
						cipher,
						compressionAlgo,
						level,
						rsa,
					)
				}
			}
		}
	}
}

func main() {
	//_ = generateEntity()
	draftBenchmarks()
}
