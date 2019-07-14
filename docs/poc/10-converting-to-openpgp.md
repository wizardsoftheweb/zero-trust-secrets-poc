# 10 Learning OpenGPG

My biggest pain point right now is `gpg`. I've got to `exec` out of everything to get to it, it makes containerization difficult and bloated, and its API is not great. Doing it directly in Go will make my life easier. Now that I have something to demo, I can make parts of it better.

Go has [a fully-featured OpenPGP library](https://godoc.org/golang.org/x/crypto/openpgp) (as well as [plenty of other neat stuff](https://godoc.org/golang.org/x/crypto)). Getting up and running with OpenPGP isn't that hard. There's [a very simple method](https://godoc.org/golang.org/x/crypto/openpgp#NewEntity) to get a fresh abstract keyring. Making a secure one quickly isn't so simple.

**NOTE:** For the most part, when I use `OpenPGP`, I'm referring to the Go library, not the full standard. I try to refer to that in other ways.

## Benchmarking Creation

I threw together a very simple and relatively secure profile to begin with. `AES256` for a symmetric cipher, zlib [just because](https://stackoverflow.com/a/20765054/2877698), the `BestSpeed` compression level, 4096 bits for RSA, and the default `SHA256` hash. That's a profile I'd be comfortable using every day. The PoC clients I'd like to build, though, can't. The time to generate fluctuates between near one second and near four seconds. That's nuts.  

### Background

OpenPGP provides [five cipher functions](https://godoc.org/golang.org/x/crypto/openpgp/packet#CipherFunction) right now, taken from [a larger list](https://www.iana.org/assignments/pgp-parameters/pgp-parameters.xhtml#pgp-parameters-13). It doesn't have some of the fancier stuff yet but [it covers the spec's requirements](https://tools.ietf.org/html/rfc4880#section-9.2).

The library enables [three compression algorithms](https://godoc.org/golang.org/x/crypto/openpgp/packet#CompressionAlgo), leaving out [BZ2, the spec's last defined compression algorithm](https://tools.ietf.org/html/rfc4880#section-9.3). Furthermore, it provides [four levels of named compression](https://github.com/golang/crypto/blob/master/openpgp/packet/compressed.go#L22) out of [ten levels total](https://github.com/golang/crypto/blob/master/openpgp/packet/compressed.go#L32).

OpenPGP can use any number of bits for RSA. [Its default is 2048](https://godoc.org/golang.org/x/crypto/openpgp/packet#Config).

Finally, [the library](https://godoc.org/golang.org/x/crypto/openpgp/packet#Config) delegates hashing to [another package](https://godoc.org/crypto#Hash) with everything you might want. [The spec](https://tools.ietf.org/html/rfc4880#section-9.4) restricts things a bit and kicks out MD5.

Finally, while there are [many allowed hash algorithms](https://tools.ietf.org/html/rfc4880#section-9.4), the library sticks with [`SHA256`](https://godoc.org/golang.org/x/crypto/openpgp/packet#Config)

### Defining the Domain

* Ciphers:
    * 3DES
    * CAST5
    * AES128
    * AES192
    * AES256
* Compression techniques:
    * None
    * ZIP
    * ZLIB
* Easy compression levels:
    * NoCompression
    * BestSpeed
    * BestCompression
    * DefaultCompression
* RSA bits:
    * 2048
    * 4096
* Hash algorithms:
    * SHA224
    * SHA256
    * SHA384
    * SHA512

This puts us at
```text
    5 ciphers
    3 compression algorithms
    4 easy compression levels
    2 bit lengths
 x  4 hash algorithms
-----------------------------
  480 Permutations 
```

### First Pass

I initially wrote a bunch of benchmarks using all the permutations but that didn't go so well. The information was very useful but it's not easily mungeable.

```shell-session
$ cat benchmark.log \
    | awk '/^PASS.*Benchmark/{ \
        split(gensub(/CreationSuite\.BenchmarkCreationOf/, "", "g", $3), group, "x"); \
        print $5 / 10 ^ 9, group[1], group[2], group[3], group[4]; }' \
    | sort -k1 -n \
    | grep 'AES256 ZLIB'
    
0.275695 AES256 ZLIB DefaultCompression 2048
0.281401 AES256 ZLIB BestSpeed 2048
0.309848 AES256 ZLIB NoCompression 2048
0.316922 AES256 ZLIB BestCompression 2048
1.3851 AES256 ZLIB DefaultCompression 4096
1.78562 AES256 ZLIB BestSpeed 4096
3.44046 AES256 ZLIB NoCompression 4096
4.38238 AES256 ZLIB BestCompression 4096
```
There's a lot of annoying work just to get the data in usable form.

### Actual Pass

Instead of fighting an uphill battle, I wrote a quick benchmark tool that writes out to a CSV for parsing. Assuming I don't move it later, it's [viewable here](/keys-from-scratch/benchmark-keygen). It's just a bunch of data and loops. If you see something I can do to make it better, let me know!

I stored the data in a CSV because it's an easy filetype to share between go routines. You can add to it without worrying about what's already there. It's also not a complicated write. Here's a small sample.
```csv
duration,hash,cipher,compAlgo,compLevel,rsa
884796935,SHA224,AES128,None,BestCompression,2048
9491934849,SHA256,3DES,None,BestSpeed,4096
4627713298,SHA384,AES128,None,BestSpeed,4096
812300430,SHA224,AES128,None,BestCompression,2048
1939383322,SHA224,AES256,None,BestCompression,2048
950667819,SHA512,3DES,None,BestCompression,2048
769944070,SHA256,3DES,None,BestCompression,2048
12228898917,SHA384,3DES,None,BestSpeed,4096
```

However, like the log file, munging a CSV is no fun. I've been playing a lot with `jq` recently and found [a solid solution that already exists](https://stackoverflow.com/a/45888945/2877698)

```shell-session
$ jq -s -R \
    '[
        [ split("\n")[] | split(",") ] \
        | { \
            h:["runtime","hash","cipher","compression","level","rsa"], \
            v:.[1:][] \
           } \
        | [.h, (.v|map(tonumber?//.))] \
        | [transpose[] \
        | {key:.[0],value:.[1]}] \
        | from_entries \
    ]' data-to-munge.csv > data-to-munge.json
[
  {
    "runtime": 21297510160,
    "hash": "SHA224",
    "cipher": "AES192",
    "compression": "None",
    "level": "BestCompression",
    "rsa": 4096
  },
  {
    "runtime": 10824945138,
    "hash": "SHA512",
    "cipher": "AES192",
    "compression": "None",
    "level": "DefaultCompression",
    "rsa": 4096
  }
]

$ jq --arg NANO "$((10**9))" \
    '[ \
        .[] \
        | .runtime = .runtime / ($NANO | tonumber) \
    ]' data-to-munge.json > divided-data.json
[
  {
    "runtime": 20.235348704,
    "hash": "SHA512",
    "cipher": "AES128",
    "compression": "None",
    "level": "BestCompression",
    "rsa": 4096
  },
  {
    "runtime": 0.987810184,
    "hash": "SHA224",
    "cipher": "AES128",
    "compression": "None",
    "level": "DefaultCompression",
    "rsa": 2048
  }
]
```
