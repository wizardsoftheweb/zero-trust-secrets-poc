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

