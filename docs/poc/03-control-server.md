# 03 Control Server Round One

I want to slow down for a minute and define the scenario I'm thinking about. On one hand, it's useful to have done criteria. On the other hand, I'd also like to document and lay out what I'm attempting to accomplish for a few reasons.

1) If it's been done before I wanna know about it and snag that FOSS.
2) If it hasn't been done for good reason, I should probably know that.
3) Sometimes I do really stupid stuff and need someone to call me out.

## The Setting

The big picture has actually exploded in scope since I sat down and started this. It's a fascinating idea. But I want to knock out a solid MVP before getting lost in all the neat stuff that can come out of this.

I am going to build a microservice that does exactly these things. Some of them might seem rather silly but each one serves to show what it can be expanded to do later.

1) Is a webserver running [Gin](https://github.com/gin-gonic/gin)
2) Exposes [Prometheus metrics](https://github.com/zsais/go-gin-prometheus)
3) Generates random, 47 character strings
4) Listens for requests that provide a KV storage address, a key to fill, and the number of 47 characters string to drop in that key
5) Fulfills those responses by either...
    * encrypting the contents in such a way that only the receivers can get to the secret, or
    * signing the response in such a way that, while it might be intercepted and inspected, it cannot be altered.

I'm really interesting in messing with Gin, so I'm not going to architect anything further now. This section is called Round One for a reason. I'm building the publisher right now but I haven't defined what motivates the publisher. I also haven't defined how subscribers interact.

Stretch goal is [messaging in proto3](https://developers.google.com/protocol-buffers/docs/proto3) instead of JSON. 

## What This is Not (Yet)

Every time I spend a few years away from serious crypto topics, the first project I do that even remotely touches crypto seems to fall into a pretty major trap. For example, I often forget that assymmetric encryption doesn't mean that two parties have two disparate secrets and trade messages with zero trust. They actually have to trade keys. During the course of writing this spec, I realized that I'm treading dangerously close to sugggesting I can implement some sort of homomorphic encryption architecture. That's what was in my head, anyway, until reality kicked in and I remembered mixing a bunch of paint just [turns everything brown](https://security.stackexchange.com/a/60659). This is not a solution for that. I'm not smart enough to even tackle the edge cases there. But if some else wants to build me a library I'd happily use it to further this goal!

## Containerizing

Because this is an MVP, I decided to forgo setting up a proper pipeline. I'm just using some image [some person published on GitHub](https://github.com/chemidy/smallest-secured-golang-docker-image). That means fix that before you call this a production solution.


## Current Code

The control server has three endpoints:

1. `/metrics`: This exposes Prometheus metrics
2. `/ping`: Boilerplate Gin; used for testing
3. `/rando`: The meat of the control server

### `/rando`

This endpoint listens for `POST` requests containing the following keys:

* `count`: The number of random strings to generate. I tested with small numbers. Obviously large numbers will make things difficult.
* `kv_hosts`: This is a list of `etcd` hosts. Might be a cluster? I didn't dig too deep; my local instance has a single endpoint, `127.0.0.1:2379`.
* `kv_key`: The key to store the encrypted random strings under
* `pub_key`: The contents of the PGP pub key

It returns the generated strings

### Example

Assume I have a control server running at `localhost:8080` with an `etcd` endpoint at `127.0.0.1:2379/`.

1. Get the contents of the pub key
    ```shell-session
    $ cat .pubring.gpg | awk '{ printf "%s\\n", $0; }'
    pub\nkey\ncontents
    ```
2. `POST` to `localhost:8080/rando`

    request:
    ```json
    {
        "count": 10,
        "kv_hosts": ["http://127.0.0.1:2379/"],
        "kv_key": "/test/value",
        "pub_key": "pub\nkey\ncontents"
    }
    ```
    
    response:
    ```json
    {
    	"message": [
    		"WGRY6KWlc-p1AD0IfI6iukkupzC834FEeOYw9Icrmm__2mi_y9wPc4pOaNWP6xE=",
    		"WMcM4leIq4enKlhYmBOWSyS4E-EjioWpN0miIUuJ38_bKBKdAz9-Q93Qsk_csWE=",
    		"7kDkFQ7ENTXrEXnpzcIbyIhgkL4AEIHvPIbvSaM8ifjKQzBPNhcmIJslZkwO-NY=",
    		"bZljvk6L44oWSbBrJHTnV7bU-IovuElq1JU8GiHxr6jZTRXo4qOaLInXakeQvOI=",
    		"ocufjMOa4rzkuxMn3Mh9YweS4Qy4BQITcgKcQ233V_uW-J5D8gLfav00cuxTBqs=",
    		"NQzoBMQtHC-OBCAHTtWJIqyDxHPRyBQhJQYUE7o-EReuj9_zFYpUjKoze3vdHo8=",
    		"4F6_jFVItGBo_ODeXKzOZYQ_cBuXvmbPtg2SS9x3JE2WW4dVeweTUY1vmFQV-CA=",
    		"q4RYskDY0M-sDt2zZebdSmZzmluKIrBmkCGzW4KbjdnjFtHeM2gxifCKZf9_L6c=",
    		"OVN2pLlwn5fC7t_d3C9gsY6xwZe6z5KHxH8V9i1dwsrAeNjYoIbXVHlF_kiXkkg=",
    		"DL9v-sVoacQA3Pqf3vjb38Av3LXlmA5ncslMBSgTb6vDz9dqGjmxa6M9KSrSEkU="
    	]
    }
    ```
3. Verify the contents of the key

    ```shell-session
    $ crypt get -endpoint='http://127.0.0.1:2379/' /test/value | sed -e $'s/,/,\\\n/g'
    ["WGRY6KWlc-p1AD0IfI6iukkupzC834FEeOYw9Icrmm__2mi_y9wPc4pOaNWP6xE=",
    "WMcM4leIq4enKlhYmBOWSyS4E-EjioWpN0miIUuJ38_bKBKdAz9-Q93Qsk_csWE=",
    "7kDkFQ7ENTXrEXnpzcIbyIhgkL4AEIHvPIbvSaM8ifjKQzBPNhcmIJslZkwO-NY=",
    "bZljvk6L44oWSbBrJHTnV7bU-IovuElq1JU8GiHxr6jZTRXo4qOaLInXakeQvOI=",
    "ocufjMOa4rzkuxMn3Mh9YweS4Qy4BQITcgKcQ233V_uW-J5D8gLfav00cuxTBqs=",
    "NQzoBMQtHC-OBCAHTtWJIqyDxHPRyBQhJQYUE7o-EReuj9_zFYpUjKoze3vdHo8=",
    "4F6_jFVItGBo_ODeXKzOZYQ_cBuXvmbPtg2SS9x3JE2WW4dVeweTUY1vmFQV-CA=",
    "q4RYskDY0M-sDt2zZebdSmZzmluKIrBmkCGzW4KbjdnjFtHeM2gxifCKZf9_L6c=",
    "OVN2pLlwn5fC7t_d3C9gsY6xwZe6z5KHxH8V9i1dwsrAeNjYoIbXVHlF_kiXkkg=",
    "DL9v-sVoacQA3Pqf3vjb38Av3LXlmA5ncslMBSgTb6vDz9dqGjmxa6M9KSrSEkU="]
    ```
### Currently Missing

* The control server does not sign its values
* The control server uses JSON not proto3
