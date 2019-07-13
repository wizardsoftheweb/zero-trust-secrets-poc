# 04 Simple Client Round One

Each client that connects to the control server needs the following information to start up:

* Location of the `etcd` servers, eg `127.0.0.1:2379`
* Root key path, eg `/simple-client-01/`

Each client that connects to the control server needs to be able to do these things:

* Create its own GPG key and export the files for use with `crypt`
* Bootstrap its secrets by hitting `/rando` on the control server
* Monitor its secrets and update them accordingly as they change

For easy proof of updated conf, I'm gonna make the client another gin server.

## Initial Configuration

This shouldn't require any explanation. The clients need to know where to store config.

## Initial Actions

### Create GPG Key and Export Files

Ideally the secret keyring should be locked down. I dunno how to address that yet. Assuming the client itself is secure, the key should be secure. Also, assuming clients are independent of each other, if one is breached, it's no big deal.

For the generation, I'm using [the basic example from `crypt`](https://github.com/xordataexchange/crypt/#create-a-key-and-keyring-from-a-batch-file) along with [my GPG export code](./02-crypt.md#create-keys).

**NOTE:** I'm not sharing config between clients yet because that's a bit more complicated. For example, this couldn't _yet_ be used to store a common database password because each client is generating that. It could, however, be used to generate a fresh user/pass for a common database, send those off to a process that can provide access to that user, and keep the password totally secret. In that case, if the client is breached, the user can be scrubbed and built again with a new client.

**NOTE:** I'm currently assuming that you've got a GPG flow. This has not been tested on a box without a GPG flow. That will come later when I containerize this. Which will probably be at the end of this doc. ¯\\\_(ツ)\_/¯

**NOTE:** This key has no passphrase. Notice the `%noprotection`. I don't have a way around that yet.

1. Check to see if the specified directory has `.pubring.gpg` and `.secring.gpg`. If both exist, done. Otherwise, continue.

2. Check to see if the GPG keyring contains a key with the note `Zero Trust Secrets` (it's not dynamic for name because this is a PoC). If it exists, skip to step 5. Else continue.

3. Create the batch file
    ```text
    %echo Generating a configuration OpenPGP key
    %no-protection
    Key-Type: default
    Subkey-Type: default
    Name-Real: CJ Harries
    Name-Comment: Zero Trust Secrets
    Name-Email: cj@wotw.pro
    Expire-Date: 0
    %commit
    %echo done
    ```
4. Run the batch file, which should put the keys in the normal key ring.

5. Determine the key ID. This might be OS-specific. I'm not sure. I only tested on Fedora 28. I think the format is constant across platforms with `gpg2`, though. If it's not, this is just a PoC and it should be easy to update.

    ```shell-session
    $ gpg2 --list-keys 'Zero Trust Secrets'
    pub   rsa2048 2019-07-13 [SC]
          something
    uid           [ultimate] CJ Harries (Zero Trust Secrets) <cj@wotw.pro>
    sub   rsa2048 2019-07-13 [E]
    ```
    
    The ID is `something` here. To parse it in `bash`, we can use `awk`.
    ```shell-session
    $ gpg2 --list-keys 'Zero Trust Secrets' | awk '/^\s/{ print $1 }'
    something
    ```
    
    However, in Go, I parse the output looking for this pattern:
    ```go
    var keyIdPattern, _ = regexp.Compile(`^\s+[^\s]*?\s*$`)
    ```
    
    If the line is a match, I strip the whitespace and return that as the ID. If I can't determine an ID, the client panics. Whoops.
    
6. Using the key ID, the client checks the existence of each file in the specified directory.

    * If `.pubring.gpg` DNE,
    
        ```shell-session
        $ gpg --output .pubring.gpg --armor --export something
        ```
        
    * If `.secring.gpg` DNE,
        
        ```shell-session
        $ gpg --output .secring.gpg --armor --export-secret-key something
        ```
        
### Bootstrap Secrets

I created a simple function to `POST` to the control server to simulate secret construction. This call requests for the local `etcd` host (`http://127.0.0.1:2379/`) to build a JSON file containing 10 strings at `/simple-client/secrets.json`. Every time the function is called, the secrets will be regenerated.

The client's boot config process is as follows:

1. Using [Viper's support for remote config](https://github.com/spf13/viper#remote-keyvalue-store-support), attempt to load `/simple-client/secrets.json` from the local `etcd` host and the provided keyring. If the file DNE, call the secret generation function.

2. With the config loaded, check the `secrets` key. If the key DNE or has no values, call the secret generation function.

3. Set `secrets` in the global state to the found config value.

You can view the current secrets by visiting `localhost:4747` once the simple client is running.

### Watch Remote Changes

Viper makes it very easy [to watch for remote changes](https://github.com/spf13/viper#watching-changes-in-etcd---unencrypted) (the example is unencrypted; encrypted is almost identical). After the client boots for the first time, Viper spins off a Go routine that runs every 30 seconds (it could be shorter but idgaf) that will update the global state. As before, you can check the current state by visiting `localhost:4747`. You can now force an update on the secrets by hitting `localhost:4747/force-update`. Once the remote watcher runs again, the secrets at `localhost:4747` should be updated.

## Putting it All Together

Ensure there's nothing in `etcd` at the desired key.
```shell-session
$ etcdctl --endpoint 'http://127.0.0.1:2379' --no-sync rm -r /simple-client
$ etcdctl --endpoint 'http://127.0.0.1:2379' --no-sync ls /simple-client
Error:  100: Key not found (/simple-client) [30]
```
Launch the control server. You can run this in the background if you aren't interested in the logs.
```shell-session
cd path/to/zero-trust-secrets/control-server
go build && ./control-server
```
Verify it's up.
```shell-session
$ curl localhost:8080/ping
{"message":"pong"}
```
Launch the simple client.
```shell-session
cd path/to/zero-trust-secrets/simple-client
go build && ./simple-client
```
Verify it's up.
```shell-session
$ curl localhost:4747/ping
{"message":"pong"}
```
Check the current secrets.
```shell-session
$ curl -s localhost:4747 | jq
{
  "secrets": [
    "DjNSv1hcDhmWw7FqR8VwQU_KYtdjftpWj3osjiDR2IFsSNjHvVMbi0j4w6-ghZs=",
    "z6ViB2tzqRRdGCyWrnd8Q9mtGo60C9Ui-bsgvQqZapc1mAMAjt9ntvfWV9BvZzk=",
    "wscoYgJVLr875HC41jfO0gCCTH3DdwRqw6Osi0lIJUdo2FyPgZlMnbSUgfxXJyY=",
    "tlKfvnjlwaZQFoQPZs5_BzLuCXiiB4OIctWvposwVQb3FbgVntE2To2atJZCsNM=",
    "voSY5B6KDEYm6b0Q-717m26UbcDV2ECG3vZfEGNU9RtBnZfo4afAwYViQG7vnWk=",
    "S0BtsHPiHCOwBEO86f2VTSLf0tRU98qv3-R3zWU9sBvydYnEzDVehr9gevd4Lyo=",
    "a_1Vsbie_trEsypEB7RvQ7Pp6BqAOuFNMAZyeawcsftVQJsTFdjxH_Inl1PpcJc=",
    "HWxX_Di5BvOWAPhhyhhp0CS4GjsyoHi18biY4riml_Y4Jc-JErU5Sf_kJ-G9XFI=",
    "FL6Ru-Efc-kQOUwr7WIDHhBfLV_m3EA4ufn1JQnQ8FUmW9eKzAsS3zYTzHQwKRI=",
    "z7sO7hq7QrdEyFHR3GuMvo5cyXtGHQlCdpjW2OuR-gKFR83c55hLCspCUSjJZ7M="
  ]
} 
```
Those secrets shouldn't change. If they do, something's wrong.
```shell-session
$ curl -s localhost:4747 | jq
{
  "secrets": [
    "DjNSv1hcDhmWw7FqR8VwQU_KYtdjftpWj3osjiDR2IFsSNjHvVMbi0j4w6-ghZs=",
    "z6ViB2tzqRRdGCyWrnd8Q9mtGo60C9Ui-bsgvQqZapc1mAMAjt9ntvfWV9BvZzk=",
    "wscoYgJVLr875HC41jfO0gCCTH3DdwRqw6Osi0lIJUdo2FyPgZlMnbSUgfxXJyY=",
    "tlKfvnjlwaZQFoQPZs5_BzLuCXiiB4OIctWvposwVQb3FbgVntE2To2atJZCsNM=",
    "voSY5B6KDEYm6b0Q-717m26UbcDV2ECG3vZfEGNU9RtBnZfo4afAwYViQG7vnWk=",
    "S0BtsHPiHCOwBEO86f2VTSLf0tRU98qv3-R3zWU9sBvydYnEzDVehr9gevd4Lyo=",
    "a_1Vsbie_trEsypEB7RvQ7Pp6BqAOuFNMAZyeawcsftVQJsTFdjxH_Inl1PpcJc=",
    "HWxX_Di5BvOWAPhhyhhp0CS4GjsyoHi18biY4riml_Y4Jc-JErU5Sf_kJ-G9XFI=",
    "FL6Ru-Efc-kQOUwr7WIDHhBfLV_m3EA4ufn1JQnQ8FUmW9eKzAsS3zYTzHQwKRI=",
    "z7sO7hq7QrdEyFHR3GuMvo5cyXtGHQlCdpjW2OuR-gKFR83c55hLCspCUSjJZ7M="
  ]
} 
```
You can force an update to change the secrets.
```shell-session
$ curl -s localhost:4747/force-update | jq
{
  "message": "Secrets were regenerated"
}
```
If you check immediately, they'll be the same.
```shell-session
$ curl -s localhost:4747 | jq; curl -s localhost:4747/force-update | jq; curl -s localhost:4747 | jq
{
  "secrets": [
    "DjNSv1hcDhmWw7FqR8VwQU_KYtdjftpWj3osjiDR2IFsSNjHvVMbi0j4w6-ghZs=",
    "z6ViB2tzqRRdGCyWrnd8Q9mtGo60C9Ui-bsgvQqZapc1mAMAjt9ntvfWV9BvZzk=",
    "wscoYgJVLr875HC41jfO0gCCTH3DdwRqw6Osi0lIJUdo2FyPgZlMnbSUgfxXJyY=",
    "tlKfvnjlwaZQFoQPZs5_BzLuCXiiB4OIctWvposwVQb3FbgVntE2To2atJZCsNM=",
    "voSY5B6KDEYm6b0Q-717m26UbcDV2ECG3vZfEGNU9RtBnZfo4afAwYViQG7vnWk=",
    "S0BtsHPiHCOwBEO86f2VTSLf0tRU98qv3-R3zWU9sBvydYnEzDVehr9gevd4Lyo=",
    "a_1Vsbie_trEsypEB7RvQ7Pp6BqAOuFNMAZyeawcsftVQJsTFdjxH_Inl1PpcJc=",
    "HWxX_Di5BvOWAPhhyhhp0CS4GjsyoHi18biY4riml_Y4Jc-JErU5Sf_kJ-G9XFI=",
    "FL6Ru-Efc-kQOUwr7WIDHhBfLV_m3EA4ufn1JQnQ8FUmW9eKzAsS3zYTzHQwKRI=",
    "z7sO7hq7QrdEyFHR3GuMvo5cyXtGHQlCdpjW2OuR-gKFR83c55hLCspCUSjJZ7M="
  ]
} 
{
  "message": "Secrets were regenerated"
}
{
  "secrets": [
    "DjNSv1hcDhmWw7FqR8VwQU_KYtdjftpWj3osjiDR2IFsSNjHvVMbi0j4w6-ghZs=",
    "z6ViB2tzqRRdGCyWrnd8Q9mtGo60C9Ui-bsgvQqZapc1mAMAjt9ntvfWV9BvZzk=",
    "wscoYgJVLr875HC41jfO0gCCTH3DdwRqw6Osi0lIJUdo2FyPgZlMnbSUgfxXJyY=",
    "tlKfvnjlwaZQFoQPZs5_BzLuCXiiB4OIctWvposwVQb3FbgVntE2To2atJZCsNM=",
    "voSY5B6KDEYm6b0Q-717m26UbcDV2ECG3vZfEGNU9RtBnZfo4afAwYViQG7vnWk=",
    "S0BtsHPiHCOwBEO86f2VTSLf0tRU98qv3-R3zWU9sBvydYnEzDVehr9gevd4Lyo=",
    "a_1Vsbie_trEsypEB7RvQ7Pp6BqAOuFNMAZyeawcsftVQJsTFdjxH_Inl1PpcJc=",
    "HWxX_Di5BvOWAPhhyhhp0CS4GjsyoHi18biY4riml_Y4Jc-JErU5Sf_kJ-G9XFI=",
    "FL6Ru-Efc-kQOUwr7WIDHhBfLV_m3EA4ufn1JQnQ8FUmW9eKzAsS3zYTzHQwKRI=",
    "z7sO7hq7QrdEyFHR3GuMvo5cyXtGHQlCdpjW2OuR-gKFR83c55hLCspCUSjJZ7M="
  ]
} 
```
However, if you wait for the refresh time, the secrets will be updated.
```shell-session
$ curl -s localhost:4747/force-update | jq; sleep 30; curl -s localhost:4747 | jq
{
  "secrets": [
    "DjNSv1hcDhmWw7FqR8VwQU_KYtdjftpWj3osjiDR2IFsSNjHvVMbi0j4w6-ghZs=",
    "z6ViB2tzqRRdGCyWrnd8Q9mtGo60C9Ui-bsgvQqZapc1mAMAjt9ntvfWV9BvZzk=",
    "wscoYgJVLr875HC41jfO0gCCTH3DdwRqw6Osi0lIJUdo2FyPgZlMnbSUgfxXJyY=",
    "tlKfvnjlwaZQFoQPZs5_BzLuCXiiB4OIctWvposwVQb3FbgVntE2To2atJZCsNM=",
    "voSY5B6KDEYm6b0Q-717m26UbcDV2ECG3vZfEGNU9RtBnZfo4afAwYViQG7vnWk=",
    "S0BtsHPiHCOwBEO86f2VTSLf0tRU98qv3-R3zWU9sBvydYnEzDVehr9gevd4Lyo=",
    "a_1Vsbie_trEsypEB7RvQ7Pp6BqAOuFNMAZyeawcsftVQJsTFdjxH_Inl1PpcJc=",
    "HWxX_Di5BvOWAPhhyhhp0CS4GjsyoHi18biY4riml_Y4Jc-JErU5Sf_kJ-G9XFI=",
    "FL6Ru-Efc-kQOUwr7WIDHhBfLV_m3EA4ufn1JQnQ8FUmW9eKzAsS3zYTzHQwKRI=",
    "z7sO7hq7QrdEyFHR3GuMvo5cyXtGHQlCdpjW2OuR-gKFR83c55hLCspCUSjJZ7M="
  ]
} 
{
  "message": "Secrets were regenerated"
}
{
  "secrets": [
    "Oy0iaGHcLH1dax7D2Rm6eT9GNuoS-11IQHXqYFCp7uwMhJgxYlpQa_BxOD7HzPY=",
    "ynlQUE-NMH3Ki6G_pVXI8dpzYzSMWC-hOVNBxhPTrwy1roAW_5qoZbdVGyPBgFc=",
    "w3XatsLyFnBnPfclMP7ALBRB8rqRbw-3xD0ZdZR8b6AuLIhylcAn4sLnjkWbOnA=",
    "OOGpZ0iU7-9N9_YZlxKwseTkq6NkeJGqObchUmDZkL5xLfl5_mBmFQMWNNL6spc=",
    "AFLsLiRQOTrs5pOodWErc6DoX-8zQbUf37tOgc-64AxhGU3kauQ0k289VogV6nc=",
    "8DpPDlGpKsSr9HIG2I6SjYO48BW5pMeDnTlk-lLrHq2NJFp7Frj1yiUQKrp3qtk=",
    "CBNfm9lByS0l_HwTNx5fCu5fo_cFpY6wH7HsbJ2kU6mex7uiUtMwdHbKq3EK9C0=",
    "uOfDYhtxQwuPaa__82oKbu7v5MpSfr6C3joH8ISyryxVaQ3ZcVdJgLUouRG8Ls0=",
    "4eM_69g2hBH3xOFZrgiEpTYNVEsBGMG8_ddHQnEwgHSPIcHmV1BVW_kAcxVZfSE=",
    "hk285R_kLA-7cMZCNQHbRGiX7PYaOPcMJ7BezPaAyK8kCm4vU-dAeg_vk-ca4YQ="
  ]
}
```
