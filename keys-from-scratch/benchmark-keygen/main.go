package main

import (
	"crypto"
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/common/log"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

//nolint:unused
const (
	nameForBenchmarks    = "Rick James"
	commentForBenchmarks = "ZTS Benchmark"
	emailForBenchmarks   = "rick.james@wotw.pro"
	repetitionCount      = 10
	maxConsectureFailure = 10
)

var (
	cipherFuncs = []packet.CipherFunction{
		packet.Cipher3DES,
		packet.CipherCAST5,
		packet.CipherAES128,
		packet.CipherAES192,
		packet.CipherAES256,
	}
	cipherNames = []string{
		"3DES",
		"CAST5",
		"AES128",
		"AES192",
		"AES256",
	}
	compressionAlgos = []packet.CompressionAlgo{
		packet.CompressionNone,
		packet.CompressionZIP,
		packet.CompressionZLIB,
	}
	compressionAlgoNames = []string{
		"None",
		"ZIP",
		"ZLIB",
	}
	compressionLevels = []int{
		packet.NoCompression,
		packet.BestSpeed,
		packet.BestCompression,
		packet.DefaultCompression,
	}
	compressionLevelNames = []string{
		"NoCompression",
		"BestSpeed",
		"BestCompression",
		"DefaultCompression",
	}
	rsaBits = []int{
		2048,
		4096,
	}
	rsaBitNames = []string{
		"2048",
		"4096",
	}
	hashFuncs = []crypto.Hash{
		crypto.SHA224,
		crypto.SHA256,
		crypto.SHA384,
		crypto.SHA512,
	}
	hashFuncNames = []string{
		"SHA224",
		"SHA256",
		"SHA384",
		"SHA512",
	}
)

func generateKey(
	hashIndex,
	cipherIndex,
	compressionAlgoIndex,
	levelIndex,
	rsaIndex int,
) (time.Duration, error) {
	start := time.Now()
	_, err := openpgp.NewEntity(
		nameForBenchmarks,
		commentForBenchmarks,
		emailForBenchmarks,
		&packet.Config{
			DefaultHash:            hashFuncs[hashIndex],
			DefaultCipher:          cipherFuncs[cipherIndex],
			DefaultCompressionAlgo: compressionAlgos[compressionAlgoIndex],
			CompressionConfig: &packet.CompressionConfig{
				Level: compressionLevels[levelIndex],
			},
			RSABits: rsaBits[rsaIndex],
		},
	)
	if nil != err {
		return time.Duration(0), err
	}
	end := time.Now()
	return end.Sub(start), nil
}

func generateKeys(logger *DataLogger, group *sync.WaitGroup, hashIndex, cipherIndex int) {
	for compressionAlgoIndex := 0; compressionAlgoIndex < len(compressionAlgos); compressionAlgoIndex++ {
		for levelIndex := 0; levelIndex < len(compressionLevels); levelIndex++ {
			for rsaIndex := 0; rsaIndex < len(rsaBits); rsaIndex++ {
				runsLeft := repetitionCount
				currentErrorCount := 0
				for 0 < runsLeft {
					duration, err := generateKey(hashIndex, cipherIndex, compressionAlgoIndex, levelIndex, rsaIndex)
					if nil != err {
						if maxConsectureFailure > currentErrorCount {
							currentErrorCount++
							continue
						} else {
							log.Fatal(err)
						}
					} else {
						err := logTimes(
							logger,
							[]string{
								strconv.FormatInt(duration.Nanoseconds(), 10),
								hashFuncNames[hashIndex],
								cipherNames[cipherIndex],
								compressionAlgoNames[compressionAlgoIndex],
								compressionLevelNames[levelIndex],
								rsaBitNames[rsaIndex],
							},
							group,
						)
						if nil != err && maxConsectureFailure > currentErrorCount {
							currentErrorCount++
							continue
						}
						currentErrorCount = 0
						runsLeft--
					}
				}
			}
		}
	}
}

func logTimes(logger *DataLogger, row []string, group *sync.WaitGroup) error {
	err := logger.Log(row)
	if nil != err {
		return err
	}
	group.Done()
	return nil
}

func main() {
	dataLogger, err := NewDataLogger("./duration.csv")
	if nil != err {
		log.Fatal(err)
	}
	rowCount := repetitionCount * len(cipherFuncs) * len(compressionAlgos) * len(compressionLevels) * len(rsaBits)
	group := &sync.WaitGroup{}
	group.Add(rowCount)
	for hashIndex := 0; hashIndex < len(hashFuncs); hashIndex++ {
		for cipherIndex := 0; cipherIndex < len(cipherFuncs); cipherIndex++ {
			go generateKeys(dataLogger, group, hashIndex, cipherIndex)
		}
	}
	group.Wait()
	dataLogger.Flush()
}
