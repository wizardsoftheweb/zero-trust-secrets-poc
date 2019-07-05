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


