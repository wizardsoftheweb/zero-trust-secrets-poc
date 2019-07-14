package main

import (
	"testing"

	"github.com/prometheus/common/log"
	"golang.org/x/crypto/openpgp"

	"golang.org/x/crypto/openpgp/packet"
)

const (
	nameForBenchmarks    = "Rick James"          //nolint:unused
	commentForBenchmarks = "ZTS Benchmark"       //nolint:unused
	emailForBenchmarks   = "rick.james@wotw.pro" //nolint:unused
)

var (
	possibleCiphersForBenchmarks = []packet.CipherFunction{ //nolint:unused
		packet.CipherAES128,
		packet.CipherAES192,
		packet.CipherAES256,
	}
	possibleCompressionAlgosForBenchmarks = []packet.CompressionAlgo{ //nolint:unused
		packet.CompressionZLIB,
		packet.CompressionZIP,
		packet.CompressionNone,
	}
	possibleCompressionLevelsForBenchmarks = []int{ //nolint:unused
		packet.BestSpeed,
		packet.BestCompression,
		packet.DefaultCompression,
		packet.NoCompression,
	}
	possibleRsaBitsForBenchmarks = []int{2048, 4096} //nolint:unused
)

var storingEntitiesToMitigateCompilerTricks *openpgp.Entity //nolint:unused

func getSingleConfigBenchmark( //nolint:unused
	cipherFunc packet.CipherFunction,
	compressionAlgo packet.CompressionAlgo,
	compressionLevel,
	rsaBits int,
	b *testing.B,
) {
	var entity *openpgp.Entity
	var err error
	for n := 0; n < b.N; n++ {
		entity, err = openpgp.NewEntity(
			nameForBenchmarks,
			commentForBenchmarks,
			emailForBenchmarks,
			&packet.Config{
				DefaultCipher:          cipherFunc,
				DefaultCompressionAlgo: compressionAlgo,
				CompressionConfig: &packet.CompressionConfig{
					Level: compressionLevel,
				},
				RSABits: rsaBits,
			},
		)
		if nil != err {
			log.Fatal("whoops")
		}
		storingEntitiesToMitigateCompilerTricks = entity
	}
}

func benchmarkCreation(b *testing.B) { //nolint:unused,deadcode
	for _, cipher := range possibleCiphersForBenchmarks {
		for _, compressionAlgo := range possibleCompressionAlgosForBenchmarks {
			for _, level := range possibleCompressionLevelsForBenchmarks {
				for _, rsa := range possibleRsaBitsForBenchmarks {
					getSingleConfigBenchmark(
						cipher,
						compressionAlgo,
						level,
						rsa,
						b,
					)
				}
			}
		}
	}
}
