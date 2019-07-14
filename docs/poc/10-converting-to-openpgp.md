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

## Just Kidding
