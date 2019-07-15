package main

import (
	"testing"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type BaseSuite struct{}

var _ = Suite(&BaseSuite{})

func (s *CreationSuite) TestForProperGenericConfig(c *C) {
	config := newGenericConfig()
	c.Assert(config.DefaultCipher, Equals, baseCipher)
	c.Assert(config.DefaultCompressionAlgo, Equals, commonCompression)
	c.Assert(config.CompressionConfig.Level, Equals, baseLevel)
	c.Assert(config.RSABits, Equals, minimumRsaBits)
}
