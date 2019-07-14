# 02 Crypt and etcd

## Installing

I'm interesting primarily in [crypt as a library](https://github.com/xordataexchange/crypt/tree/master/config) but it will be faster to figure out via the CLI.

```shell-session
go get github.com/xordataexchange/crypt/bin/crypt
go install github.com/xordataexchange/crypt/bin/crypt
```

## Making a Key

A quick note before you keep reading. I'm sharing basically every step of my MVP key gen process which means, you guessed it, these keys are worthless now. The process is great but only if you keep it to yourself.

I'm running a fairly recent version with plenty of algorithms so I'm going to experiment a little bit with what I can use.

```shell-session
$ gpg2 --version
gpg (GnuPG) 2.2.8
libgcrypt 1.8.3
Copyright (C) 2018 Free Software Foundation, Inc.
License GPLv3+: GNU GPL version 3 or later <https://gnu.org/licenses/gpl.html>
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Home: /home/cjharries/.gnupg
Supported algorithms:
Pubkey: RSA, ELG, DSA, ECDH, ECDSA, EDDSA
Cipher: IDEA, 3DES, CAST5, BLOWFISH, AES, AES192, AES256, TWOFISH,
        CAMELLIA128, CAMELLIA192, CAMELLIA256
Hash: SHA1, RIPEMD160, SHA256, SHA384, SHA512, SHA224
Compression: Uncompressed, ZIP, ZLIB, BZIP2

$ gpg2 --full-gen-key
gpg (GnuPG) 2.2.8; Copyright (C) 2018 Free Software Foundation, Inc.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Please select what kind of key you want:
   (1) RSA and RSA (default)
   (2) DSA and Elgamal
   (3) DSA (sign only)
   (4) RSA (sign only)
Your selection?
```
However, in order to do that, we'll need to escalate our perms.
```
$ man gpg2 | grep -A 8 -- --expert
--expert
--no-expert
      Allow the user to do certain nonsensical or "silly" things  like
      signing an expired or revoked key, or certain potentially incom‐
      patible things like generating unusual key types. This also dis‐
      ables  certain  warning  messages about potentially incompatible
      actions. As the name implies, this option is for  experts  only.
      If you don't fully understand the implications of what it allows
      you to do, leave this off. --no-expert disables this option.
```
As fully qualified experts, having read the docs, we can test out some new algos. I'm still going to do a traditional RSA and RSA key. However, if the services all run with an ed25519 key I would be in heaven.
```
$ gpg2 --expert --full-gen-key
gpg (GnuPG) 2.2.8; Copyright (C) 2018 Free Software Foundation, Inc.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Please select what kind of key you want:
   (1) RSA and RSA (default)
   (2) DSA and Elgamal
   (3) DSA (sign only)
   (4) RSA (sign only)
   (7) DSA (set your own capabilities)
   (8) RSA (set your own capabilities)
   (9) ECC and ECC
  (10) ECC (sign only)
  (11) ECC (set your own capabilities)
  (13) Existing key
Your selection? 9
Please select which elliptic curve you want:
   (1) Curve 25519
   (3) NIST P-256
   (4) NIST P-384
   (5) NIST P-521
   (9) secp256k1
Your selection? 1
...
$ gpg2 --full-gen-key
gpg (GnuPG) 2.2.8; Copyright (C) 2018 Free Software Foundation, Inc.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.

Please select what kind of key you want:
   (1) RSA and RSA (default)
   (2) DSA and Elgamal
   (3) DSA (sign only)
   (4) RSA (sign only)
Your selection? 1
RSA keys may be between 1024 and 4096 bits long.
What keysize do you want? (2048)
Requested keysize is 2048 bits

$ gpg2 -k zero-trust-configuration
pub   ed25519 2019-07-05 [SC]
      DC3D0941E8BD4F616D543944633CA5EE6A2CD95C
uid           [ultimate] CJ Harries (zero-trust-configuration) <cj@wizardsoftheweb.pro>
sub   cv25519 2019-07-05 [E]

pub   rsa2048 2019-07-05 [SC]
      BA4F78AE1B4760C795E0ACBCB456DC07E4F1BF74
uid           [ultimate] CJ Harries (zero-trust-configuration) <cj@wizardsoftheweb.pro>
sub   rsa2048 2019-07-05 [E]
```

I'm gonna pop these off my keyring so I can just keep testing with the same ones and nuke them when I'm done.

```
# These have to be done with gpg2
$ gpg2 --armor --export-secret-keys -a DC3D0941E8BD4F616D543944633CA5EE6A2CD95C > .ed25519.asc
$ gpg2 --armor --export -a DC3D0941E8BD4F616D543944633CA5EE6A2CD95C > .ed25519_pub.asc
# Nothing special
$ gpg --armor --export-secret-keys -a BA4F78AE1B4760C795E0ACBCB456DC07E4F1BF74 > .rsa.asc
$ gpg --armor --export -a BA4F78AE1B4760C795E0ACBCB456DC07E4F1BF74 > .rsa_pub.asc
$ exa
.rw-rw-r--@  784 cjharries 06:31 .ed25519.asc
.rw-rw-r--@  681 cjharries 06:32 .ed25519_pub.asc
.rw-rw-r--@ 3.6k cjharries 06:32 .rsa.asc
.rw-rw-r--@ 1.8k cjharries 06:32 .rsa_pub.asc
```

## Using the Client

Let's start by pulling the entry I made creating etcd.
```shell-session
$ ENDPOINT=$(systemctl status etcd | grep -Eo 'http://localhost:[0-9]+' | uniq)
$ crypt get -backend=etcd \
            -endpoint=$ENDPOINT \
            -secret-keyring=./.rsa.gpg \
            /my/first/key
```
### That's a Nope

Turns out generating the keys the way I did doesn't work with crypt. I think. I'm currently at a loss because [the way they did it](https://github.com/xordataexchange/crypt#create-a-key-and-keyring-from-a-batch-file) to set up the project (mind you, that was like five years ago) has [been no-oped](https://www.gnupg.org/documentation/manuals/gnupg/Unattended-GPG-key-generation.html) out of the batch files.

I did get something to either work or break. I'm not sure what yet. If I can't figure it out quickly enough I'll go to my fallback which involves more k8s heavy lifting that I was hoping for.

### The Resolution

The library crypt relies on for etcd has been deprecated [for some time](https://github.com/coreos/go-etcd). It's not a great situation to be in but it's not bad either. The codebase is small enough that it would be quick to patch.

The CLI does connect without encryption, so that's a start!

```shell-session
$ crypt get -plaintext -endpoint="http://127.0.0.1:2379" /my/first/key
my-first-value
```

That success led me to investigate a bit more. I think it didn't work due a combination of touchy code, bad documentation, and, most importantly, my own inexperience. The first thing I discovered looking through the codebase was the the arguments had defaults. Once I matched the defaults I didn't have any trouble. Sometimes that's all it takes for you to admit you screwed something up and learned from the process!

## Once More From the Top

### Install crypt
```shell-session
go get github.com/xordataexchange/crypt/bin/crypt
go install github.com/xordataexchange/crypt/bin/crypt
```

### Create Keys

Check the output of this.
```shell-session
$ man gpg2 | grep -E -A 1 -- "^\s*?--secret-keyring"
       --secret-keyring file
              This is an obsolete option and ignored...
```
If it looks like that, you can't use [the original solution](https://github.com/xordataexchange/crypt#generating-gpg-keys-and-keyrings). You could also check your version of `gpg2`; it dropped in 2.1. It's just more fun to comb the `man` pages.

If, like me, you're unable to take the easy batch route, just do this instead. You'll need basic keys, not for any lack of trying on crypt's part. [Their backend](https://github.com/golang/crypto/blob/a49355c7e3f8fe157a85be2f77e6e269a0f89602/openpgp/packet/public_key_v3.go#L69) currently only supports RSA.

```shell-session
$ gpg2 --full-key-gen
...
$ gpg --output .secring.gpg --armor --export-secret-key BA4F78AE1B4760C795E0ACBCB456DC07E4F1BF74
$ gpg --output .pubring.gpg --armor --export BA4F78AE1B4760C795E0ACBCB456DC07E4F1BF74
```

### Confirm etcd Addresses

Parsing the config file, `/etc/etcd/etcd.conf`, and parsing SystemD should give you similar results. You're looking for the client addresses.
```shell-session
$ systemctl status etcd | grep -oE http://localhost:\[0-9\]+ | uniq
http://localhost:2379
```

### Use crypt to Query etcd

```shell-session
# etcdctl setup
$ echo '{"test":"value"}' > config.json
$ etcdctl --no-sync mkdir /demo
$ etcdctl --no-sync set /demo/etcdctl/json "$(cat ./config.json)"
{"test":"value"}

# crypt setup
$ export URL='http://127.0.0.1:2379'
$ test -f .secring.gpg
$ test -f .pubring.gpg
$ crypt set -endpoint=$URL /demo/crypt/json config.json

# getting with etcdctl
$ etcdctl --no-sync get /demo/etcdctl/json
{"test":"value"}

$ etcdctl --no-sync get /demo/crypt/json
wcBMA+qYkqI+MRHiAQgAf9YAQidQ2uUieUG6ft4zGWAOchGttsuUo2/QjLHaMvz/pbHOKg7Iig24Tp7fUXW3gnBhaaM57rfiT93C5hEm89UI2B/Vh1GRJqrOS8OKQDvhiG+v4dYtw8rVzpRe4/Dq7WWvKJy0nctzFuKz8hscEhob0x5EmDPxrG8Hon79P9EVUDYCfRMLTFulAufZ5QJ1HlnmD3p40tDuMCQouNaJSLEk6YFHIRjfHXFVwxnAjCFYTLDiUFhhmMFAw/R4RLrhSJYiFjxZ/0pUcVsjFwXwxURspI94fc/LSN7Xm4aXpSKeFoq/lEbC8/Txqtnm7o4ZPcVCJ8elBOd8bN4OaWLV89LgAeRcUunMZj5StijcxVJPS5Uf4XRu4DTgaeEjOeCL4kVApIzgFOMskvvJrYG8f+AT4SAm4C7ku/3CgVDDm3+PgJZuUYY5deBI4Tf24HHgu+B94s/ndMXg3eO5CCtwQcmSnODW5OQwFdBRAA4CiLDV7QZm4rDiXfcEFeF5lQA=

# getting with crypt
$ crypt get -endpoint=$URL /demo/etcdctl/json
illegal base64 data at input byte 0

$ crypt get -endpoint=$URL /demo/crypt/json
{"test":"value"}

# listing with etcdctl
$ etcdctl --no-sync ls /demo
/demo/etcdctl
/demo/crypt

$ etcdctl --no-sync ls /demo/etcdctl
/demo/etcdctl/json

$ etcdctl --no-sync ls /demo/crypt
/demo/crypt/json

# listing with crypt
$ crypt list -endpoint=$URL /demo

$ crypt list -endpoint=$URL /demo/etcdctl
illegal base64 data at input byte 0

$ crypt list -endpoint=$URL /demo/crypt
/demo/crypt/json: {"test":"value"}
```
## Conclusion

This is neat. This is very neat. It's very basic and doesn't do much more than encrypting data before throwing it somewhere, but that's what I'm trying to accomplish! I'm going to investigate Viper next. I feel that it should handily address crypt's QoL deficiencies.
