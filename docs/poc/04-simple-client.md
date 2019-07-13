# 04 Simple Client Round One

Each client that connects to the control server needs the following information to start up:

* Location of the `etcd` servers, eg `127.0.0.1:2379`
* Root key path, eg `/simple-client-01/`

Each client that connects to the control server needs to be able to do these things:

* Create its own GPG key and export the files for use with `crypt`
* Bootstrap its secrets by hitting `/rando` on the control server
* Monitor its secrets and update them accordingly as they change

## Initial Configuration

This shouldn't require any explanation. The clients need to know where to store config.

## Initial Actions

### Create GPG Key and Export Files

Ideally the secret keyring should be locked down. I dunno how to address that yet. Assuming the client itself is secure, the key should be secure. Also, assuming clients are independent of each other, if one is breached, it's no big deal.

**NOTE:** I'm not sharing config between clients yet because that's a bit more complicated. For example, this couldn't _yet_ be used to store a common database password because each client is generating that. It could, however, be used to generate a fresh user/pass for a common database, send those off to a process that can provide access to that user, and keep the password totally secret. In that case, if the client is breached, the user can be scrubbed and built again with a new client.

For the generation, I'm using [the basic example from `crypt`](https://github.com/xordataexchange/crypt/#create-a-key-and-keyring-from-a-batch-file) along with [my GPG export code](./02-crypt.md#create-keys).

**NOTE:** I'm currently assuming that you've got a GPG flow. This has not been tested on a box without a GPG flow. That will come later when I containerize this. Which will probably be at the end of this doc. ¯\_(ツ)_/¯

**NOTE:** This key has no passphrase. Notice the `%noprotection`. I don't have a way around that yet.

```text
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
```


