package main

import (
	"math/rand"
	"time"

	"github.com/prometheus/common/log"
	"golang.org/x/crypto/openpgp"
	. "gopkg.in/check.v1"

	"golang.org/x/crypto/openpgp/packet"
)

//nolint:unused
const (
	nameForBenchmarks    = "Rick James"
	commentForBenchmarks = "ZTS Benchmark"
	emailForBenchmarks   = "rick.james@wotw.pro"
)

func seedEnv() bool { //nolint:unused
	rand.Seed(time.Now().UnixNano())
	return true
}

//nolint:unused,deadcode,varcheck
var (
	isEnvSeeded                             = seedEnv()
	storingEntitiesToMitigateCompilerTricks *openpgp.Entity
)

type CreationSuite struct {
}

var _ = Suite(&CreationSuite{})

func (s *CreationSuite) TestForProperGenericConfig(c *C) {
	config := newGenericConfig()
	c.Assert(config.DefaultCipher, Equals, baseCipher)
	c.Assert(config.DefaultCompressionAlgo, Equals, commonCompression)
	c.Assert(config.CompressionConfig.Level, Equals, baseLevel)
	c.Assert(config.RSABits, Equals, baseLevel)
}

func benchmarkSingleConfig( //nolint:unused
	cipherFunc packet.CipherFunction,
	compressionAlgo packet.CompressionAlgo,
	compressionLevel,
	rsaBits int,
	c *C,
) {
	var entity *openpgp.Entity
	var err error
	for n := 0; n < c.N; n++ {
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

func (s *CreationSuite) BenchmarkCreationOfAES128xZLIBxBestSpeedx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZLIB, packet.BestSpeed, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZLIBxBestSpeedx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZLIB, packet.BestSpeed, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZLIBxBestCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZLIB, packet.BestCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZLIBxBestCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZLIB, packet.BestCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZLIBxDefaultCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZLIB, packet.DefaultCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZLIBxDefaultCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZLIB, packet.DefaultCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZLIBxNoCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZLIB, packet.NoCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZLIBxNoCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZLIB, packet.NoCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZIPxBestSpeedx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZIP, packet.BestSpeed, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZIPxBestSpeedx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZIP, packet.BestSpeed, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZIPxBestCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZIP, packet.BestCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZIPxBestCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZIP, packet.BestCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZIPxDefaultCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZIP, packet.DefaultCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZIPxDefaultCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZIP, packet.DefaultCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZIPxNoCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZIP, packet.NoCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xZIPxNoCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionZIP, packet.NoCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xNonexBestSpeedx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionNone, packet.BestSpeed, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xNonexBestSpeedx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionNone, packet.BestSpeed, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xNonexBestCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionNone, packet.BestCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xNonexBestCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionNone, packet.BestCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xNonexDefaultCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionNone, packet.DefaultCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xNonexDefaultCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionNone, packet.DefaultCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xNonexNoCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionNone, packet.NoCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES128xNonexNoCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES128, packet.CompressionNone, packet.NoCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZLIBxBestSpeedx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZLIB, packet.BestSpeed, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZLIBxBestSpeedx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZLIB, packet.BestSpeed, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZLIBxBestCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZLIB, packet.BestCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZLIBxBestCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZLIB, packet.BestCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZLIBxDefaultCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZLIB, packet.DefaultCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZLIBxDefaultCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZLIB, packet.DefaultCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZLIBxNoCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZLIB, packet.NoCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZLIBxNoCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZLIB, packet.NoCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZIPxBestSpeedx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZIP, packet.BestSpeed, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZIPxBestSpeedx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZIP, packet.BestSpeed, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZIPxBestCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZIP, packet.BestCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZIPxBestCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZIP, packet.BestCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZIPxDefaultCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZIP, packet.DefaultCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZIPxDefaultCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZIP, packet.DefaultCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZIPxNoCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZIP, packet.NoCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xZIPxNoCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionZIP, packet.NoCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xNonexBestSpeedx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionNone, packet.BestSpeed, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xNonexBestSpeedx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionNone, packet.BestSpeed, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xNonexBestCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionNone, packet.BestCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xNonexBestCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionNone, packet.BestCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xNonexDefaultCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionNone, packet.DefaultCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xNonexDefaultCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionNone, packet.DefaultCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xNonexNoCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionNone, packet.NoCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES192xNonexNoCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES192, packet.CompressionNone, packet.NoCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZLIBxBestSpeedx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZLIB, packet.BestSpeed, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZLIBxBestSpeedx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZLIB, packet.BestSpeed, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZLIBxBestCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZLIB, packet.BestCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZLIBxBestCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZLIB, packet.BestCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZLIBxDefaultCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZLIB, packet.DefaultCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZLIBxDefaultCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZLIB, packet.DefaultCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZLIBxNoCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZLIB, packet.NoCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZLIBxNoCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZLIB, packet.NoCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZIPxBestSpeedx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZIP, packet.BestSpeed, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZIPxBestSpeedx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZIP, packet.BestSpeed, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZIPxBestCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZIP, packet.BestCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZIPxBestCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZIP, packet.BestCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZIPxDefaultCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZIP, packet.DefaultCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZIPxDefaultCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZIP, packet.DefaultCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZIPxNoCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZIP, packet.NoCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xZIPxNoCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionZIP, packet.NoCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xNonexBestSpeedx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionNone, packet.BestSpeed, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xNonexBestSpeedx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionNone, packet.BestSpeed, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xNonexBestCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionNone, packet.BestCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xNonexBestCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionNone, packet.BestCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xNonexDefaultCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionNone, packet.DefaultCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xNonexDefaultCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionNone, packet.DefaultCompression, 4096, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xNonexNoCompressionx2048(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionNone, packet.NoCompression, 2048, c)
}

func (s *CreationSuite) BenchmarkCreationOfAES256xNonexNoCompressionx4096(c *C) {
	benchmarkSingleConfig(packet.CipherAES256, packet.CompressionNone, packet.NoCompression, 4096, c)
}
