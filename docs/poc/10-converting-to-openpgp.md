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

### `jq` Fun

I got really frustrated with `jq-1.6` and its reliance on some `C` libraries that apparently play differently enough on RHEL systems that it won't work. There's a chance I tried every permutation from [the README's instructions](https://github.com/stedolan/jq/blob/master/README.md) without any luck. I also followed the instructions on [the `jq` site](https://stedolan.github.io/jq/download/#from_source_on_linux_os_x_cygwin_and_other_posixlike_operating_systems) but had the same problems. Using the binary release doesn't work either. Using the release from the Fedora repos doesn't work either. All I wanted to do was calculate a power!

```shell-session
$ jq 'pow(10,2)'
jq: error: pow/1 is not defined at <top-level>, line 1:
pow(10, 2)
jq: 1 compile error
```

If you were paying attention above, I got around it for a little bit by using `bash` math. That's unpleasant and not a great pipeline. So I switched to R.

### This Time It's Actual

Once I decided to switch to R, the whole need to convert from CSV to JSON went away. R makes [this kind of ETL very easy](https://idc9.github.io/stor390/notes/dplyr/dplyr.html). I didn't initially use R because I didn't want to deal with yet another GUI or console. But while I was researching `jq` solutions, I ran into [`Rscript`](https://support.rstudio.com/hc/en-us/articles/218012917-How-to-run-R-scripts-from-the-command-line) which is brilliant and I don't know why I didn't think to research that before. If I had ever bothered to take the time and learn [the `pandas` flow](https://pandas.pydata.org/pandas-docs/stable/getting_started/comparison/comparison_with_r.html) I probably wouldn't have wasted so much time poking around with `jq`. That being said, I'm stoked because R is super easy to use.

The full script is [here](/keys-from-scratch/benchmark-keygen/munge.R). I'd love feedback! I haven't touched R in forever. I graduated and moved away from R before [the `tidyverse` flow](https://github.com/tidyverse/tidyverse) showed up. Hadley Wickham has done some [really rad stuff](https://r4ds.had.co.nz/) with R in the years that I've been not doing math. If you see something that could be improved with the R stuff, let me know!

The first thing I did was get the CSV loaded and sorted. I reordered the table and converted nanoseconds to seconds for easy consumption.

```text
    RSA Cipher   Hash Compr     Level  Duration
1  4096 AES128 SHA224  None BestCompr 2.5726776
2  2048 AES256 SHA512  None   BestSpd 0.2482865
3  4096   3DES SHA384   ZIP BestCompr 6.7017688
4  4096 AES128 SHA256  ZLIB BestCompr 1.6331067
5  4096   3DES SHA224  ZLIB BestCompr 2.8825446
6  4096   3DES SHA256  None   BestSpd 8.4028635
7  2048 AES256 SHA256   ZIP   BestSpd 0.3520052
8  4096 AES256 SHA512  None   BestSpd 3.7405410
9  2048 AES256 SHA512  None  DefCompr 0.2594043
10 4096  CAST5 SHA384  ZLIB  DefCompr 3.3297039
```

I got a little concerned because there's an insane amount of variability between different configs. To make sure I wasn't getting bad samples, I spent a lot of time tweaking how I was generating the CSV. Originally I was creating too many go routines, which gave me a lot of outliers and skewed my results toward longer processes. Cutting down the number of go routines greatly increased the total generation, unfortunately. It did give me better data, though.

```text
9600 records / 480 permutations = 20 records per permutation
```

Admittedly, that's not a lot. But I lost patience and started investigating anyway. Initially, I looked at all the results together, grouping by everything not `Duration` so I could get a median duration for each permutation.
```text
     Hash Cipher Compr     Level  RSA medianDur meanDur durationSd coeffOfVar
22 SHA512  CAST5  ZLIB   NoCompr 2048    0.3655  0.3660    0.08264      22.58
23 SHA512 AES256  ZLIB   BestSpd 2048    0.3955  0.4218    0.14950      35.44
24 SHA224 AES256  None BestCompr 2048    0.3309  0.3402    0.11920      35.05
25 SHA224   3DES   ZIP   NoCompr 4096    4.2190  4.4420    1.84300      41.50
26 SHA224  CAST5  None BestCompr 4096    3.2430  3.2790    1.35400      41.30
27 SHA512 AES256   ZIP   NoCompr 4096    2.0980  3.1380    1.76800      56.35
28 SHA256  CAST5  ZLIB BestCompr 2048    0.3897  0.3693    0.09691      26.24
29 SHA384 AES192   ZIP   NoCompr 2048    0.3725  0.3952    0.16290      41.22
30 SHA224 AES128  ZLIB   BestSpd 4096    3.3240  4.5850    3.17100      69.16
31 SHA512 AES128  ZLIB   NoCompr 4096    2.8150  3.4080    1.65800      48.65
```
As you can see, I ended up calculating more because the median was not helpful. It's all over the place. In many cases, though, it's close enough to the mean to suggest some grouping. I've forgotten most of my stats, so I just ran with [the coefficient of variation](https://en.wikipedia.org/wiki/Coefficient_of_variation) as a check. It's
```text
(sample standard deviation / mean) * 100 := unitless percentage
```
The smaller the coefficient is, the less volatile the data. Volatile data, or a high variance, greatly reduces your ability to model and predict. In other words, it's pretty weak and can't be used as a good benchmark because it's way too random.

I started doing some secondary searches to get a feel for what was going on. An easy target was the data where there's not compression algorithm. I didn't originally catch it, but, in theory, if the compression is gone, the compression level shouldn't affect anything. I [searched the codebase](https://github.com/golang/crypto/search?q=compression+path%3Aopenpgp&unscoped_q=compression+path%3Aopenpgp) to confirm that; it looks legit. I played with a few other groupings.

* `Compr==None` grouping `Level`
* `Compr==None` without grouping `Level`
* grouping everything except for `Level` and `Duration`
* grouping everything except for `Compr`. `Level`, and `Duration`

The chart below shows the min, max, mean, and median of each set's coefficient of variation.

```text
            name      min      max     mean   median
1  ComprNoneWLvl 23.22994 70.68649 40.43760 39.34945
2 ComprNoneWoLvl 32.43989 55.05657 41.37174 40.90893
3         SumLvl 31.26976 55.07023 42.08036 42.15585
4    SumLvlCompr 34.31455 54.02505 42.27768 42.47947
5            All 22.31983 73.04754 41.27713 40.92455
```
While some of the filters tightened the range, everything stayed around 40%, which is a lot of volatility. I decided to look at what data was doing what to see if that would lend any clues. I began by looking at values whose median duration was under five seconds that also had a coefficient of variation below 50%. That filter reduced the search space from 480 down to a whopping 414. I did notice a fairly obvious pattern, but it's fairly obvious because it was fairly obvious before this whole project started.
```text
     Hash Cipher Compr     Level  RSA medianDur meanDur durationSd coeffOfVar
1  SHA224 AES128   ZIP   NoCompr 2048    0.3649  0.3482    0.07772      22.32
2  SHA512  CAST5  ZLIB   NoCompr 2048    0.3655  0.3660    0.08264      22.58
3  SHA256 AES256  None   NoCompr 2048    0.4540  0.4521    0.10500      23.23
4  SHA224 AES128  ZLIB  DefCompr 4096    3.4770  3.3700    0.81520      24.19
5  SHA512 AES192  None   NoCompr 2048    0.3152  0.3498    0.08544      24.43
6  SHA256 AES192   ZIP   BestSpd 4096    3.0510  2.8920    0.71220      24.63
7  SHA512 AES192  ZLIB   BestSpd 2048    0.3435  0.3442    0.08550      24.84
8  SHA512 AES128   ZIP  DefCompr 2048    0.3216  0.3266    0.08208      25.13
9  SHA224 AES256   ZIP   BestSpd 2048    0.3500  0.3402    0.08694      25.55
10 SHA256  CAST5  ZLIB BestCompr 2048    0.3897  0.3693    0.09691      26.24
11 SHA256 AES256   ZIP BestCompr 2048    0.3500  0.3572    0.09488      26.56
12 SHA256   3DES   ZIP   NoCompr 2048    0.4073  0.4202    0.11280      26.84
13 SHA384 AES192  ZLIB  DefCompr 2048    0.4343  0.4173    0.11300      27.09
14 SHA512 AES192  ZLIB BestCompr 2048    0.3249  0.3395    0.09275      27.32
15 SHA224 AES256   ZIP BestCompr 2048    0.3597  0.3622    0.10050      27.75
16 SHA384 AES192  ZLIB   BestSpd 2048    0.3867  0.3822    0.10680      27.95
17 SHA384   3DES  None BestCompr 2048    0.3379  0.3509    0.09905      28.22
18 SHA384  CAST5  None   NoCompr 2048    0.3238  0.3244    0.09167      28.26
19 SHA256 AES256  ZLIB BestCompr 2048    0.4047  0.4045    0.11480      28.39
20 SHA384 AES128  ZLIB   BestSpd 2048    0.3459  0.3301    0.09417      28.53
```
It's significantly slower to use more bits. Larger numbers are harder to work with. Here, it's around 700% slower.
```text
                     min      max   mean   median
medianDur % inc 419.1293 1425.708 735.49 723.3816
```
I'd personally like to run with more bits because it's more secure. That really depends on how slow the fastest 4096 keys are being made. For an ideal key, I'd like to see 4096 bits, a very quick gen time with a hard cap at five seconds, and some compression (either zip or zlib at some level above zero). For me to trust the prediction, the coefficient of variation should be below 40%. That filter gave me 41 results.
```text
     Hash Cipher Compr     Level  RSA medianDur meanDur durationSd coeffOfVar
1  SHA512  CAST5  ZLIB BestCompr 4096     2.257   2.322     0.8883      38.26
2  SHA384  CAST5   ZIP  DefCompr 4096     2.412   2.752     0.9587      34.84
3  SHA224 AES256   ZIP  DefCompr 4096     2.501   2.568     0.8582      33.42
4  SHA512   3DES   ZIP   BestSpd 4096     2.591   2.637     0.9358      35.48
5  SHA224 AES128   ZIP BestCompr 4096     2.653   2.933     1.1300      38.53
6  SHA512   3DES  ZLIB   BestSpd 4096     2.733   2.922     0.8866      30.34
7  SHA224 AES192   ZIP  DefCompr 4096     2.752   3.092     1.1130      35.98
8  SHA384  CAST5  ZLIB   BestSpd 4096     2.764   2.799     1.0740      38.37
9  SHA256  CAST5   ZIP   BestSpd 4096     2.794   2.821     0.9141      32.40
10 SHA224  CAST5   ZIP   BestSpd 4096     2.818   2.915     1.0660      36.56
```
Sorted by median generation, the fastest configs take more than two seconds. That's nuts.
```console
$ bash -c 'trap times EXIT; gpg2 --armor --batch --gen-key gpg.batch'
gpg: Generating a configuration OpenPGP key
gpg: key A341F25CE1CF7CAA marked as ultimately trusted
gpg: revocation certificate stored as '~/.gnupg/openpgp-revocs.d/B8E01B6233B41C705255DDCAA341F25CE1CF7CAA.rev'
gpg: done
0m0.004s 0m0.002s
0m0.010s 0m0.000s
```
`gpg2` takes milliseconds to do almost the same task. That doesn't make any sense.

### Back to Solo Key Gen

I decided to take the config from that first row and try messing with it. There's a chance generating keys via go routines instead of the main process is affecting it. Except that's not how Go works. The main process is basically a go routine with focus. Either way, I wanted to generate it [another way](/keys-from-scratch/main.go) to see what would happen.

```console
$ go build && ./keys-from-scratch
2019/07/14 17:35:54 duration: 2.410806551s
2019/07/14 17:35:54 &openpgp.Entity{PrimaryKey:(*packet.PublicKey)(0xc0004d2000), PrivateKey:(*packet.PrivateKey)(0xc0004d4000), Identities:map[string]*openpgp.Identity{"CJ Harries (cj@wotw.pro) <Home Brew ZTS PoC>":(*openpgp.Identity)(0xc000030080)}, Revocations:[]*packet.Signature(nil), Subkeys:[]openpgp.Subkey{openpgp.Subkey{PublicKey:(*packet.PublicKey)(0xc0004d2280), PrivateKey:(*packet.PrivateKey)(0xc0004d41a0), Sig:(*packet.Signature)(0xc0004d61c0)}}}

$ ./keys-from-scratch
2019/07/14 17:37:27 duration: 1.542727705s
2019/07/14 17:37:27 &openpgp.Entity{PrimaryKey:(*packet.PublicKey)(0xc0002f8000), PrivateKey:(*packet.PrivateKey)(0xc0002fa000), Identities:map[string]*openpgp.Identity{"CJ Harries (cj@wotw.pro) <Home Brew ZTS PoC>":(*openpgp.Identity)(0xc00020c080)}, Revocations:[]*packet.Signature(nil), Subkeys:[]openpgp.Subkey{openpgp.Subkey{PublicKey:(*packet.PublicKey)(0xc0002f8280), PrivateKey:(*packet.PrivateKey)(0xc0002fa1a0), Sig:(*packet.Signature)(0xc0002fc1c0)}}}

$ ./keys-from-scratch
2019/07/14 17:37:36 duration: 3.257900837s
2019/07/14 17:37:36 &openpgp.Entity{PrimaryKey:(*packet.PublicKey)(0xc0002a4000), PrivateKey:(*packet.PrivateKey)(0xc0002a6000), Identities:map[string]*openpgp.Identity{"CJ Harries (cj@wotw.pro) <Home Brew ZTS PoC>":(*openpgp.Identity)(0xc000204040)}, Revocations:[]*packet.Signature(nil), Subkeys:[]openpgp.Subkey{openpgp.Subkey{PublicKey:(*packet.PublicKey)(0xc0002a4280), PrivateKey:(*packet.PrivateKey)(0xc0002a61a0), Sig:(*packet.Signature)(0xc0002a81c0)}}}

$ ./keys-from-scratch
2019/07/14 17:38:43 duration: 1.465451441s
2019/07/14 17:38:43 &openpgp.Entity{PrimaryKey:(*packet.PublicKey)(0xc0004d4000), PrivateKey:(*packet.PrivateKey)(0xc0004d6000), Identities:map[string]*openpgp.Identity{"CJ Harries (cj@wotw.pro) <Home Brew ZTS PoC>":(*openpgp.Identity)(0xc000030080)}, Revocations:[]*packet.Signature(nil), Subkeys:[]openpgp.Subkey{openpgp.Subkey{PublicKey:(*packet.PublicKey)(0xc0004d4280), PrivateKey:(*packet.PrivateKey)(0xc0004d61a0), Sig:(*packet.Signature)(0xc0004d81c0)}}}

$ ./keys-from-scratch
2019/07/14 17:39:24 duration: 3.712138163s
2019/07/14 17:39:24 &openpgp.Entity{PrimaryKey:(*packet.PublicKey)(0xc0004e4000), PrivateKey:(*packet.PrivateKey)(0xc0004e6000), Identities:map[string]*openpgp.Identity{"CJ Harries (cj@wotw.pro) <Home Brew ZTS PoC>":(*openpgp.Identity)(0xc000250140)}, Revocations:[]*packet.Signature(nil), Subkeys:[]openpgp.Subkey{openpgp.Subkey{PublicKey:(*packet.PublicKey)(0xc0004e4280), PrivateKey:(*packet.PrivateKey)(0xc0004e61a0), Sig:(*packet.Signature)(0xc0004e81c0)}}}
```
This doesn't make any sense. 2.5 seconds is the median with the rest of the set split equally a second away. A two second range for something that should be a rote process has to mean something else is going.

### `pprof` to the Rescue?

If you're not familiar with `pprof`, neither was I until this problem. It's hella useful, though. It's [super fast to set up for simple processes](https://flaviocopes.com/golang-profiling/) and, assuming you played with it locally first to get a handle on it, [equally fast to use with long-running tasks](https://jvns.ca/blog/2017/09/24/profiling-go-with-pprof/).

Once it's inserted in the codebase, it's good to go.

```console
$ go build && ./keys-from-scratch
2019/07/14 17:48:39 profile: cpu profiling enabled, ~/zero-trust-secrets/keys-from-scratch/cpu.pprof
2019/07/14 17:48:44 duration: 5.427781549s
2019/07/14 17:48:44 &openpgp.Entity{PrimaryKey:(*packet.PublicKey)(0xc0002b2000), PrivateKey:(*packet.PrivateKey)(0xc0002b4000), Identities:map[string]*openpgp.Identity{"CJ Harries (cj@wotw.pro) <Home Brew ZTS PoC>":(*openpgp.Identity)(0xc000030080)}, Revocations:[]*packet.Signature(nil), Subkeys:[]openpgp.Subkey{openpgp.Subkey{PublicKey:(*packet.PublicKey)(0xc0002b2280), PrivateKey:(*packet.PrivateKey)(0xc0002b41a0), Sig:(*packet.Signature)(0xc0002b61c0)}}}
2019/07/14 17:48:44 profile: cpu profiling disabled, ~/zero-trust-secrets/keys-from-scratch/cpu.pprof
```
What this should show us is what happened during this obscenely long call. If it's something I did (ie did wrong), it should be as obvious as the time jump from doubling bits.

**NOTE:** You don't actually need to use a double dash for the flags. I do it because I'm opinionated. Short flags are for stacking options. Long flags are for human-readable content and more complicated things. [`pprof` accepts both](https://github.com/google/pprof/blob/e84dfd68c163c45ea47aa24b3dc7eaa93f6675b1/internal/driver/interactive.go#L302) just like [BSD's implementation](https://www.freebsd.org/cgi/man.cgi?getopt_long(3)) and [GNU's implementation](https://linux.die.net/man/3/getopt_long). Doesn't mean I have to like it.

```console
$ go tool pprof --text ./keys-from-scratch ./cpu.pprof
File: keys-from-scratch
Type: cpu
Time: Jul 14, 2019 at 5:48pm (CDT)
Duration: 5.60s, Total samples = 5.40s (96.35%)
Showing nodes accounting for 5.34s, 98.89% of 5.40s total
Dropped 33 nodes (cum <= 0.03s)
      flat  flat%   sum%        cum   cum%
     4.20s 77.78% 77.78%      4.20s 77.78%  math/big.addMulVVW
     0.85s 15.74% 93.52%      5.31s 98.33%  math/big.nat.montgomery
     0.27s  5.00% 98.52%      0.27s  5.00%  runtime.memclrNoHeapPointers
     0.02s  0.37% 98.89%      0.03s  0.56%  math/big.nat.divLarge
         0     0% 98.89%      5.38s 99.63%  crypto/rand.Prime
         0     0% 98.89%      5.38s 99.63%  crypto/rsa.GenerateKey
         0     0% 98.89%      5.38s 99.63%  crypto/rsa.GenerateMultiPrimeKey
         0     0% 98.89%      5.39s 99.81%  golang.org/x/crypto/openpgp.NewEntity
         0     0% 98.89%      5.39s 99.81%  main.generateEntity
         0     0% 98.89%      5.39s 99.81%  main.main
         0     0% 98.89%      5.38s 99.63%  math/big.(*Int).ProbablyPrime
         0     0% 98.89%      0.03s  0.56%  math/big.basicSqr
         0     0% 98.89%      0.27s  5.00%  math/big.nat.clear
         0     0% 98.89%      0.03s  0.56%  math/big.nat.div
         0     0% 98.89%      5.31s 98.33%  math/big.nat.expNN
         0     0% 98.89%      5.31s 98.33%  math/big.nat.expNNMontgomery
         0     0% 98.89%      0.06s  1.11%  math/big.nat.probablyPrimeLucas
         0     0% 98.89%      5.32s 98.52%  math/big.nat.probablyPrimeMillerRabin
         0     0% 98.89%      0.03s  0.56%  math/big.nat.sqr
         0     0% 98.89%      5.39s 99.81%  runtime.main
```
I like doing things via the CLI. For visual readers, I also [generated a PNG](/keys-from-scratch/debug.png). It's massive.
```console
$ go tool pprof  --png ./keys-from-scratch ./cpu.pprof > debug.png
```
Let's break down what this is saying.

* `Duration: 5.60s, Total samples = 5.40s (96.35%)`: The samples Go tool cover the solid majority of the execution so we can be fairly certain the issue is covered here.
* `Showing nodes accounting for 5.34s, 98.89% of 5.40s total`: Go's showing us the important stuff that covers our window of interest. There are some things it's not showing us but that's okay.
* The `flat` column shows us how much time was spent at each node.
* The `cum` column, while out of order, shows us the acculumated time at each node.

You can use other outputs, such as the tree or a graph, to get a visual path through the process. I really like `--tree`, which is much larger and visual, `--text` seemed like the best option for a quick overview (although if you scroll to the end of this doc you'll find my absolute favorite).

This not-so-little guy seems to be the problem.
```console
$ go tool pprof  --text --compact_labels --show 'addMulVVW' ./keys-from-scratch ./cpu.pprof
Active filters:
   show=addMulVVW
Showing nodes accounting for 4.20s, 77.78% of 5.40s total
      flat  flat%   sum%        cum   cum%
     4.20s 77.78% 77.78%      4.20s 77.78%  math/big.addMulVVW

$ ./keys-from-scratch >/dev/null 2>&1 && go tool pprof  --text --compact_labels --show 'addMulVVW' ./keys-from-scratch ./cpu.pprof
Active filters:
   show=addMulVVW
Showing nodes accounting for 3.96s, 72.13% of 5.49s total
      flat  flat%   sum%        cum   cum%
     3.96s 72.13% 72.13%      3.96s 72.13%  math/big.addMulVVW

$ ./keys-from-scratch >/dev/null 2>&1 && go tool pprof  --text --compact_labels --show 'addMulVVW' ./keys-from-scratch ./cpu.pprof
Active filters:
   show=addMulVVW
Showing nodes accounting for 3.52s, 75.70% of 4.65s total
      flat  flat%   sum%        cum   cum%
     3.52s 75.70% 75.70%      3.52s 75.70%  math/big.addMulVVW
```
It's sitting at a fairly stable 75%. It's not the whole issue but it seems to be most of it.
```console
# lowered to 2048 bits
$ go build && ./keys-from-scratch >/dev/null 2>&1 && go tool pprof  --text --compact_labels --show 'addMulVVW' ./keys-from-scratch ./cpu.pprof
Active filters:
   show=addMulVVW
Showing nodes accounting for 370ms, 67.27% of 550ms total
      flat  flat%   sum%        cum   cum%
     370ms 67.27% 67.27%      370ms 67.27%  math/big.addMulVVW

# dropped to SHA256
$ go build && ./keys-from-scratch >/dev/null 2>&1 && go tool pprof  --text --compact_labels --show 'addMulVVW' ./keys-from-scratch ./cpu.pprof
Active filters:
   show=addMulVVW
Showing nodes accounting for 140ms, 60.87% of 230m**s total
      flat  flat%   sum%        cum   cum%
     140ms 60.87% 60.87%      140ms 60.87%  math/big.addMulVVW

# bumped cipher up to AES258
$ go build && ./keys-from-scratch >/dev/null 2>&1 && go tool pprof  --text --compact_labels --show 'addMulVVW' ./keys-from-scratch ./cpu.pprof
Active filters:
   show=addMulVVW
Showing nodes accounting for 160ms, 61.54% of 260ms total
      flat  flat%   sum%        cum   cum%
     160ms 61.54% 61.54%      160ms 61.54%  math/big.addMulVVW

# bumped hash up to SHA512
$ go build && ./keys-from-scratch >/dev/null 2>&1 && go tool pprof  --text --compact_labels --show 'addMulVVW' ./keys-from-scratch ./cpu.pprof
Active filters:
   show=addMulVVW
Showing nodes accounting for 80ms, 80.00% of 100ms total
      flat  flat%   sum%        cum   cum%
      80ms 80.00% 80.00%       80ms 80.00%  math/big.addMulVVW
```

## Final Thoughts

It turns out that this a known issue with Go.

* [Same thing in 2014](https://grokbase.com/t/gg/golang-nuts/14cbkyv5kf/go-nuts-slow-math-big-performance-and-the-impact-on-crypto-tls)
* [Two years ago](https://github.com/golang/go/issues/22643)
* [Last year](https://github.com/containous/traefik/issues/2673#issuecomment-374381116)

There's [been work on it](https://go-review.googlesource.com/q/addMulVVW). It's either still embarrassingly slow and they're just avoiding the issue or there's been a regression. [This change suggests the former](https://go-review.googlesource.com/c/go/+/164966). The final comments in that change create [this issue](https://github.com/golang/go/issues/32492) and say it's on the agenda for Go 1.14.

In other words, I didn't accomplish anything I wanted to because the underlying library is horrible. That would be very frustrating if I didn't get to do a ton of really cool stuff. The PoC isn't where I want it but it's still doing well.

There are a few potential leads that I plan on investigating when I'm able to come back to this.

1) The recommended replacement for `math/big` is [a C-based library, GMP](https://github.com/ncw/gmp). I saw [a note in the official Go repo](https://github.com/golang/go/issues/22643) and in several of the other issues I dug up. An external C library might solve the speed issue but it opens up a slew of new ones.
2) There's a chance it's just my box. Lots of the issues had lots of people saying `It works fine on my machine`. It might work differently containerized on `arm64`. I know [I can build `arm64`](https://medium.com/@kurt.stam/building-aarch64-arm-containers-on-dockerhub-d2d7c975215c) on my `amd64` box and I'm pretty sure [I can run it too](https://blog.hypriot.com/post/docker-intel-runs-arm-containers/), so it's worth a shot.
3) I never went past initial key generation. There's a chance things get better after that. A slow key build is fine if the it's nowhere near as slow in use. I highly doubt this one but it should be fairly simple to cross off.
4) I could lower my standards.
5) There are other encryption schemes. They'll require some additional work to set up with QoL tools like `etcd` and Viper. Having seen the innards of `crypt` (and `etcdctl`), I don't think adding that functionality would be very difficult.

I've learned about a ton of different products today alone in trying to figure out this Golang issue. I'm sad I didn't get further along with programmatically generating keys, but I think the PoC is in a good place.

## `pprof` Easter Egg
I moved this to the end so it wouldn't be bothersome to people using Markdown renderers that don't parse HTML. I think it's hilarious and I'm going to try to use [`graph-easy`](https://github.com/ironcamel/Graph-Easy) everywhere now.

<details>
<summary>You might find this obnoxious but I love it</summary>
<p>

```console
$ go tool pprof --dot ./keys-from-scratch ./cpu.pprof | graph-easy --ascii
                                                                + - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
                                                                '                           cluster_L                           '
                                                                '                                                               '
                                                                ' +-----------------------------------------------------------+ '
                                                                ' |                  File: keys-from-scratch                  | '
                                                                ' | Type: cpu                                                 | '
                                                                ' | Time: Jul 14, 2019 at 5:48pm (CDT)                        | '
                                                                ' | Duration: 5.60s, Total samples = 5.40s (96.35%)           | '
                                                                ' | Showing nodes accounting for 5.34s, 98.89% of 5.40s total | '
                                                                ' | Dropped 33 nodes (cum <= 0.03s)                           | '
                                                                ' +-----------------------------------------------------------+ '
                                                                '                                                               '
                                                                + - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - - +
                                                                  +-----------------------------------------------------------+
                                                                  |                          runtime                          |
                                                                  |                           main                            |
                                                                  |                    0 of 5.39s (99.81%)                    |
                                                                  +-----------------------------------------------------------+
                                                                    |
                                                                    | 5.39s
                                                                    v
                                                                  +-----------------------------------------------------------+
                                                                  |                           main                            |
                                                                  |                           main                            |
                                                                  |                    0 of 5.39s (99.81%)                    |
                                                                  +-----------------------------------------------------------+
                                                                    |
                                                                    | 5.39s
                                                                    v
                                                                  +-----------------------------------------------------------+
                                                                  |                           main                            |
                                                                  |                      generateEntity                       |
                                                                  |                    0 of 5.39s (99.81%)                    |
                                                                  +-----------------------------------------------------------+
                                                                    |
                                                                    | 5.39s
                                                                    v
                                                                  +-----------------------------------------------------------+
                                                                  |                          openpgp                          |
                                                                  |                         NewEntity                         |
                                                                  |                    0 of 5.39s (99.81%)                    |
                                                                  +-----------------------------------------------------------+
                                                                    |
                                                                    | 5.38s
                                                                    v
                                                                  +-----------------------------------------------------------+
                                                                  |                            rsa                            |
                                                                  |                        GenerateKey                        |
                                                                  |                    0 of 5.38s (99.63%)                    |
                                                                  +-----------------------------------------------------------+
                                                                    |
                                                                    | 5.38s
                                                                    v
                                                                  +-----------------------------------------------------------+
                                                                  |                            rsa                            |
                                                                  |                   GenerateMultiPrimeKey                   |
                                                                  |                    0 of 5.38s (99.63%)                    |
                                                                  +-----------------------------------------------------------+
                                                                    |
                                                                    | 5.38s
                                                                    v
                                                                  +-----------------------------------------------------------+
                                                                  |                           rand                            |
                                                                  |                           Prime                           |
                                                                  |                    0 of 5.38s (99.63%)                    |
                                                                  +-----------------------------------------------------------+
                                                                    |
                                                                    | 5.38s
                                                                    v
+--------------------+          +--------------------+            +-----------------------------------------------------------+
|        big         |          |        big         |            |                            big                            |
|        nat         |          |        nat         |            |                          (*Int)                           |
|        div         |  0.03s   | probablyPrimeLucas |  0.06s     |                       ProbablyPrime                       |
| 0 of 0.03s (0.56%) | <------- | 0 of 0.06s (1.11%) | <-------   |                    0 of 5.38s (99.63%)                    |
+--------------------+          +--------------------+            +-----------------------------------------------------------+
  |                               |                                 |
  | 0.03s                         | 0.03s                           | 5.32s
  v                               v                                 v
+--------------------+          +--------------------+            +-----------------------------------------------------------+
|        big         |          |        big         |            |                            big                            |
|        nat         |          |        nat         |            |                            nat                            |
|      divLarge      |          |        sqr         |            |                 probablyPrimeMillerRabin                  |
|   0.02s (0.37%)    |          | 0 of 0.03s (0.56%) |            |                    0 of 5.32s (98.52%)                    |
|  of 0.03s (0.56%)  |          |                    |            |                                                           |
+--------------------+          +--------------------+            +-----------------------------------------------------------+
                                  |                                 |
                                  | 0.03s                           | 5.30s
                                  v                                 v
                                +--------------------+            +-----------------------------------------------------------+
                                |        big         |            |                            big                            |
                                |      basicSqr      |            |                            nat                            |
                                | 0 of 0.03s (0.56%) |            |                           expNN                           |
                                |                    |            |                    0 of 5.31s (98.33%)                    |
                                +--------------------+            +-----------------------------------------------------------+
                                  |                                 |
                                  | 0.03s                           | 5.31s
                                  v                                 v
                                +--------------------+            +-----------------------------------------------------------+
                                |        big         |            |                            big                            |
                                |     addMulVVW      |            |                            nat                            |
                                |   4.20s (77.78%)   |            |                      expNNMontgomery                      |
                                |                    |            |                    0 of 5.31s (98.33%)                    |
                                +--------------------+            +-----------------------------------------------------------+
                                  ^                                 |
                                  |                                 | 5.31s
                                  |                                 v
                                  |                               +-----------------------------------------------------------+
                                  |                               |                            big                            |
                                  |                               |                            nat                            |
                                  |                               |                        montgomery                         |
                                  |                    4.17s      |                      0.85s (15.74%)                       |
                                  +----------------------------   |                     of 5.31s (98.33%)                     |
                                                                  +-----------------------------------------------------------+
                                                                    |
                                                                    | 0.27s
                                                                    v
                                                                  +-----------------------------------------------------------+
                                                                  |                            big                            |
                                                                  |                            nat                            |
                                                                  |                           clear                           |
                                                                  |                    0 of 0.27s (5.00%)                     |
                                                                  +-----------------------------------------------------------+
                                                                    |
                                                                    | 0.27s
                                                                    v
                                                                  +-----------------------------------------------------------+
                                                                  |                          runtime                          |
                                                                  |                   memclrNoHeapPointers                    |
                                                                  |                       0.27s (5.00%)                       |
                                                                  +-----------------------------------------------------------+
```
</p>
</detail>
