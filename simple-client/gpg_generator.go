package main

const gpgBatchFile = `\
%echo Generating a configuration OpenPGP key
Key-Type: default
Subkey-Type: default
Name-Real: CJ Harries
Name-Comment: Zero Trust Secrets
Name-Email: cj@wotw.pro
Expire-Date: 0
%pubring .pubring.gpg
%commit
%echo done
`
