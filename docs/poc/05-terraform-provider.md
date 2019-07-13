# 05 Terraform

For solid adoption, this needs to work with Terraform. If you're using something else to provision your multicloud infrastructure and it works really well, please shoot me an email because I really hate TF and would love to use literally anything else that works well. Hating beta and breaking changes in production aside, TF's the right choice for any sort of infrastructure as code.

## `terraform-provider-zts`

The provider requires a control server be running.

```hcl-terraform
provider "zts" {
  control_server = "fqdn"
}
```
It provides two resources:

* `zts_secrets`: Using a predefined set of GPG keys and an existing `etcd` host, creates secrets on the `etcd` host signed with the provided keys.
* `zts_gpg_keys`: Creates the GPG keys (TODO)

### `zts_secrets`

This resource provides another way to get at the control server's `/rando` endpoint. It needs to know the `etcd` host and key to store secrets in. It assumes a common secret count and defaults to the default key ring names in the current directory unless new values are passed in. 

```hcl-terraform
resource "zts_secrets" "sample" {
  etcd_host    = "127.0.0.1:2379"            # Required
  etcd_key     = "/test-client/secrets.json" # Required
  secret_count = 10                          # Optional, default 10
  pub_key      = "/path/to/pub/key"          # Optional, default ./.pubring.gpg
  secret_key   = "/path/to/secret/key"       # Optional, default ./.secring.gpg
}
```

#### `zts_secrets` Example


```hcl-terraform
# ~/example/main.tf
provider "zts" {
  control_server = "localhost:8080"
}

resource "zts_secrets" "test" {
  etcd_host = "http://127.0.0.1:2379/"
  etcd_key  = "/test-client/secrets.json"
}

output "random_secrets" {
  value = "${zts_secrets.test.random_secrets}"
}
```

```shell-session
$ cd ~/example
$ etcdctl --endpoint 'http://127.0.0.1:2379/' --no-sync get /test-client/secrets.json 
Error:  100: Key not found (/test-client/secrets.json) [54]

$ crypt get -endpoint='http://127.0.0.1:2379' /test-client/secrets.json | jq 
100: Key not found (/test-client/secrets.json) [54]

$ terraform plan
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.


------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # zts_secrets.test will be created
  + resource "zts_secrets" "test" {
      + etcd_host      = "http://127.0.0.1:2379/"
      + etcd_key       = "/test-client/secrets.json"
      + id             = (known after apply)
      + pub_key        = "~/example/.pubring.gpg"
      + random_secrets = (known after apply)
      + secret_count   = 10
      + secret_key     = "~/example/.secring.gpg"
    }

Plan: 1 to add, 0 to change, 0 to destroy.

------------------------------------------------------------------------

Note: You didn't specify an "-out" parameter to save this plan, so Terraform
can't guarantee that exactly these actions will be performed if
"terraform apply" is subsequently run.

$ terraform apply

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # zts_secrets.test will be created
  + resource "zts_secrets" "test" {
      + etcd_host      = "http://127.0.0.1:2379/"
      + etcd_key       = "/test-client/secrets.json"
      + id             = (known after apply)
      + pub_key        = "~/example/.pubring.gpg"
      + random_secrets = (known after apply)
      + secret_count   = 10
      + secret_key     = "~/example/.secring.gpg"
    }

Plan: 1 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

zts_secrets.test: Creating...
zts_secrets.test: Creation complete after 0s [id=7b114242c3c6b5bc2490f14d337b841bc9faa424e32236ba48592714d04e0086]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:

random_secrets = [
  "0280V3PXnpgpms8_WufMXzH_K5FnIFRDcKjGc8EGR8vc71TK4UVmZQ07bP6HbEM=",
  "RdnE1lQtBVwm-GuNeMfh4pXwX1POwW2xt0kL92urcGjIM6gXEgyVijwz0d9q-yM=",
  "aWfS2DUCwvQeXztpMOjsIXK0iFJuOjDaCG6ECHX_NOdY9n00T7Gg35JFt3BxyKQ=",
  "TiWw3JY-d6yltmTCa1wbJc3LuUJtXuv5rChbVLXRvAikQCUHXlfVOlDIhJbLWOw=",
  "I4JtNsUDR1c7qUU65l9Y7Vr1FU8J-YltL2kPchlS7fVRuIIE610MMoHGR8Cbw54=",
  "-i40CzEgRU4Deh6dwMubEPcjC__b8DG_j1IU13lmxAaTzemlb7I8ezFHm-7ntO8=",
  "o3ohMsK_T9a_XY0l99sqp3QhXkMIctVGVKly5NpXMWSfCI_1CzUbxJvwaiRn1R8=",
  "8h6QHqL0diZMRsymVYndeK_wwMXkgURR-eWwX44jpccs_HqFY-EKAfsNOHWKJbQ=",
  "J7DJHGPhCMpz5x5GwvuIo4vEZw8qTzR5w2jMDASJD4Nbz6yKNWgwjpGfXL15NUc=",
  "pHQphurKiQCoP9aQBZPWFwJkxly_E0aVAG0mLWso9yAoW5WhJj5yTTh56JRlSLU=",
]

$ etcdctl --endpoint 'http://127.0.0.1:2379/' --no-sync get /test-client/secrets.json 
wcBMA9HSmdzRsOKxAQgAvbZTG3NWpstHrMYP0uzOKjn8fosdFKGTG9tg9SyLxFan3shTlPDp9OE6rNh6rmHzZ4MMePDKNcRxOWyQB196KY07HFDqUhyHkG15+6lxwbOwm40yJkcEWzQhSRaefy0lurjSktfnynkVd+Da3DHsxbgcu8x9mXA1n3VvHFMaAFSfzfIb89W+k1yZ/eTv1ZFvTdo4xuNv6y7vzOrGJkwbtzjL6Wc4v/Xg/MTsH+rRlxZ/eRfQTdiMMOUKE8pBFPpdl8gVDBNzNzmj+Bkfkouxzom6kwt3nsi0rXATuMYWzNCwZX7u7gy6DUTfd4FyYI6/C4pfxNvHCk61Ds7IU4Ew5NLgAeRW/KdFW5Bz5SXO7k5J/94G4Zyf4JXgneHLdOAf4jsqcbbggeMrx+v6hj3Oh+DL4f5/4LvnfLchBVppDarcTIl14YiCQFwurDDLzpkwwWu0uEIAGwxrDtlDDqwwPXqfiiYRbitBlMTg/iRNbKZ7VZo08y1Yl4Evf6u8st71OGBxwHks3DZ7VxsqxjLECVWIL4N/ftElXuNMXbZq6DR3LAIWI0v/E2tmOdCCKyUy2LBd7Vq/EW3gfOZA7dY7Q6VMx3X0CoJzSTKMIK9e2VyOmmpIRAB2xllSL6ym2pZ4L2B4AsfOkS36ESZ83tKaznQuiSmYz/QNW5tH4H7lytZSuoeAxmlZw9g2mJYnJYAKOczcIB6N740PI5XCmSLg5uTTUVtXHo3QFk1ptDrMsuoI4IrnsVxD5iWjez8dYHooDXFNhNGG/SKqhc521woo18HOhw08+oEFgiuziG4MwWPnBuzkOeLndWykqNygb9+0ID51VxewAnq5rdw2aNtjeeTxGvqGhlU/XKO2XxaM7yZpBDcjL1bnWLwPDnyrAGbgewKblYtKaVABt3lB7EzD8qJzpSngTuYByytV6yfzRuBU9PaElcBy7cwrM0HrwHZsQ31qzR6jikwrNw4BUW1DEyOP3drZ+vS4RW7hvGJ1QWkQqW4ow3s24NTlNFG+nCpYfrJev8WVKf8csce0CZ+MK5FYrtmNFS5p+ofgweTrYO/obFKLGplqHDCSCCR64E3mUXx1Z7rhI5gvlPcP23hnPHNDf+0/Ky5V21Zm1TOfh0F+jSwcVWy0Uf/r09szi8GxDDwYTfu58HjVGOqEKBU/0+DA4qYtLFTgD+JF69z64Jnjlp9uB5hgA8jg0OQe6u1xuJvv3wDNgcVrOfy14u5Pd+zheVAA

$ crypt get -endpoint='http://127.0.0.1:2379' /test-client/secrets.json | jq 
{
  "secrets": [
    "0280V3PXnpgpms8_WufMXzH_K5FnIFRDcKjGc8EGR8vc71TK4UVmZQ07bP6HbEM=",
    "RdnE1lQtBVwm-GuNeMfh4pXwX1POwW2xt0kL92urcGjIM6gXEgyVijwz0d9q-yM=",
    "aWfS2DUCwvQeXztpMOjsIXK0iFJuOjDaCG6ECHX_NOdY9n00T7Gg35JFt3BxyKQ=",
    "TiWw3JY-d6yltmTCa1wbJc3LuUJtXuv5rChbVLXRvAikQCUHXlfVOlDIhJbLWOw=",
    "I4JtNsUDR1c7qUU65l9Y7Vr1FU8J-YltL2kPchlS7fVRuIIE610MMoHGR8Cbw54=",
    "-i40CzEgRU4Deh6dwMubEPcjC__b8DG_j1IU13lmxAaTzemlb7I8ezFHm-7ntO8=",
    "o3ohMsK_T9a_XY0l99sqp3QhXkMIctVGVKly5NpXMWSfCI_1CzUbxJvwaiRn1R8=",
    "8h6QHqL0diZMRsymVYndeK_wwMXkgURR-eWwX44jpccs_HqFY-EKAfsNOHWKJbQ=",
    "J7DJHGPhCMpz5x5GwvuIo4vEZw8qTzR5w2jMDASJD4Nbz6yKNWgwjpGfXL15NUc=",
    "pHQphurKiQCoP9aQBZPWFwJkxly_E0aVAG0mLWso9yAoW5WhJj5yTTh56JRlSLU="
  ]
}

$ terraform refresh
zts_secrets.test: Refreshing state... [id=7b114242c3c6b5bc2490f14d337b841bc9faa424e32236ba48592714d04e0086]

Outputs:

random_secrets = [
  "0280V3PXnpgpms8_WufMXzH_K5FnIFRDcKjGc8EGR8vc71TK4UVmZQ07bP6HbEM=",
  "RdnE1lQtBVwm-GuNeMfh4pXwX1POwW2xt0kL92urcGjIM6gXEgyVijwz0d9q-yM=",
  "aWfS2DUCwvQeXztpMOjsIXK0iFJuOjDaCG6ECHX_NOdY9n00T7Gg35JFt3BxyKQ=",
  "TiWw3JY-d6yltmTCa1wbJc3LuUJtXuv5rChbVLXRvAikQCUHXlfVOlDIhJbLWOw=",
  "I4JtNsUDR1c7qUU65l9Y7Vr1FU8J-YltL2kPchlS7fVRuIIE610MMoHGR8Cbw54=",
  "-i40CzEgRU4Deh6dwMubEPcjC__b8DG_j1IU13lmxAaTzemlb7I8ezFHm-7ntO8=",
  "o3ohMsK_T9a_XY0l99sqp3QhXkMIctVGVKly5NpXMWSfCI_1CzUbxJvwaiRn1R8=",
  "8h6QHqL0diZMRsymVYndeK_wwMXkgURR-eWwX44jpccs_HqFY-EKAfsNOHWKJbQ=",
  "J7DJHGPhCMpz5x5GwvuIo4vEZw8qTzR5w2jMDASJD4Nbz6yKNWgwjpGfXL15NUc=",
  "pHQphurKiQCoP9aQBZPWFwJkxly_E0aVAG0mLWso9yAoW5WhJj5yTTh56JRlSLU=",
]

$ terraform destroy
zts_secrets.test: Refreshing state... [id=7b114242c3c6b5bc2490f14d337b841bc9faa424e32236ba48592714d04e0086]

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  # zts_secrets.test will be destroyed
  - resource "zts_secrets" "test" {
      - etcd_host      = "http://127.0.0.1:2379/" -> null
      - etcd_key       = "/test-client/secrets.json" -> null
      - id             = "7b114242c3c6b5bc2490f14d337b841bc9faa424e32236ba48592714d04e0086" -> null
      - pub_key        = "~/example/.pubring.gpg" -> null
      - random_secrets = [
          - "0280V3PXnpgpms8_WufMXzH_K5FnIFRDcKjGc8EGR8vc71TK4UVmZQ07bP6HbEM=",
          - "RdnE1lQtBVwm-GuNeMfh4pXwX1POwW2xt0kL92urcGjIM6gXEgyVijwz0d9q-yM=",
          - "aWfS2DUCwvQeXztpMOjsIXK0iFJuOjDaCG6ECHX_NOdY9n00T7Gg35JFt3BxyKQ=",
          - "TiWw3JY-d6yltmTCa1wbJc3LuUJtXuv5rChbVLXRvAikQCUHXlfVOlDIhJbLWOw=",
          - "I4JtNsUDR1c7qUU65l9Y7Vr1FU8J-YltL2kPchlS7fVRuIIE610MMoHGR8Cbw54=",
          - "-i40CzEgRU4Deh6dwMubEPcjC__b8DG_j1IU13lmxAaTzemlb7I8ezFHm-7ntO8=",
          - "o3ohMsK_T9a_XY0l99sqp3QhXkMIctVGVKly5NpXMWSfCI_1CzUbxJvwaiRn1R8=",
          - "8h6QHqL0diZMRsymVYndeK_wwMXkgURR-eWwX44jpccs_HqFY-EKAfsNOHWKJbQ=",
          - "J7DJHGPhCMpz5x5GwvuIo4vEZw8qTzR5w2jMDASJD4Nbz6yKNWgwjpGfXL15NUc=",
          - "pHQphurKiQCoP9aQBZPWFwJkxly_E0aVAG0mLWso9yAoW5WhJj5yTTh56JRlSLU=",
        ] -> null
      - secret_count   = 10 -> null
      - secret_key     = "~/example/.secring.gpg" -> null
    }

Plan: 0 to add, 0 to change, 1 to destroy.

Do you really want to destroy all resources?
  Terraform will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes

zts_secrets.test: Destroying... [id=7b114242c3c6b5bc2490f14d337b841bc9faa424e32236ba48592714d04e0086]
zts_secrets.test: Destruction complete after 0s

Destroy complete! Resources: 1 destroyed.

$ etcdctl --endpoint 'http://127.0.0.1:2379/' --no-sync get /test-client/secrets.json 
Error:  100: Key not found (/test-client/secrets.json) [56]

$ crypt get -endpoint='http://127.0.0.1:2379' /test-client/secrets.json | jq 
100: Key not found (/test-client/secrets.json) [56]
```

### `zts_gpg_keys`

This resource is probably super broken. I don't have a solid grasp on how to organize and streamline TF providers and ended up with a massive file. It works (until it doesn't) but it's far from pretty.

The resource provides another way to generate GPG keys locally. tbh I don't see where it falls in what would be normal usage but it was an interesting experiment nonetheless. It requires a name, email, and comment. It assumes files should be built in the current working directory using the normal `crypt` names (`.{pub,sec}ring.gpg`) but all those things can be changed.

```hcl-terraform
resource "zts_gpg_keys" "sample" {
  directory  = "~/example"            # Optional; default is cwd
  pub_key    = "some-pub-basename"    # Optional; default is .pubring.gpg
  secret_key = "some-secret-basename" # Optional; default is .secring.gpg
  
  batch {
    name    = "CJ Harries"            # Required
    email   = "cj@wotw.pro"           # Required
    comment = "TF ZTS"                # Required
  }
}
```

#### `zts_gpg_keys` Example

```hcl-terraform
# ~/example/main.tf
provider "zts" {
  control_server = "localhost:8080"
}

resource "zts_gpg_keys" "test" {
  batch {
    name    = "CJ Harries"
    email   = "cj@wotw.pro"
    comment = "TF ZTS"
  }
}

output "gpg_key_id" {
  value = "${zts_gpg_keys.test.gpg_key_id}"
}
```

```shell-session
$ cd ~/example
$ gpg2 --list-keys 'cj@wotw.pro'
gpg: error reading key: No public key

$ ls ./.*.gpg
zsh: no matches found: ./.*.gpg

$ terraform plan

Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.


------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # zts_gpg_keys.test will be created
  + resource "zts_gpg_keys" "test" {
      + computed_pub_key    = (known after apply)
      + computed_secret_key = (known after apply)
      + directory           = "~/example"
      + gpg_key_id          = (known after apply)
      + id                  = (known after apply)
      + pub_key             = ".pubring.gpg"
      + secret_key          = ".secring.gpg"

      + batch {
          + comment = "TF ZTS"
          + email   = "cj@wotw.pro"
          + name    = "CJ Harries"
        }
    }

Plan: 1 to add, 0 to change, 0 to destroy.

------------------------------------------------------------------------

Note: You didn't specify an "-out" parameter to save this plan, so Terraform
can't guarantee that exactly these actions will be performed if
"terraform apply" is subsequently run.

$ terraform apply
An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # zts_gpg_keys.test will be created
  + resource "zts_gpg_keys" "test" {
      + computed_pub_key    = (known after apply)
      + computed_secret_key = (known after apply)
      + directory           = "~/example"
      + gpg_key_id          = (known after apply)
      + id                  = (known after apply)
      + pub_key             = ".pubring.gpg"
      + secret_key          = ".secring.gpg"

      + batch {
          + comment = "TF ZTS"
          + email   = "cj@wotw.pro"
          + name    = "CJ Harries"
        }
    }

Plan: 1 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

zts_gpg_keys.test: Creating...
zts_gpg_keys.test: Creation complete after 1s [id=908CB26FEE8FD0AB4482F9367A511F6DDFA194B8]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:

gpg_key_id = 908CB26FEE8FD0AB4482F9367A511F6DDFA194B8

$ gpg2 --list-keys 'cj@wotw.pro'
pub   rsa2048 2019-07-13 [SC]
      908CB26FEE8FD0AB4482F9367A511F6DDFA194B8
uid           [ultimate] CJ Harries (TF ZTS) <cj@wotw.pro>
sub   rsa2048 2019-07-13 [E]

$ ls ./.*.gpg
./.pubring.gpg  ./.secring.gpg

$ terraform destroy
zts_gpg_keys.test: Refreshing state... [id=908CB26FEE8FD0AB4482F9367A511F6DDFA194B8]

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  # zts_gpg_keys.test will be destroyed
  - resource "zts_gpg_keys" "test" {
      - computed_pub_key    = "~/example/.pubring.gpg" -> null
      - computed_secret_key = "~/example/.secring.gpg" -> null
      - directory           = "~/example" -> null
      - gpg_key_id          = "908CB26FEE8FD0AB4482F9367A511F6DDFA194B8" -> null
      - id                  = "908CB26FEE8FD0AB4482F9367A511F6DDFA194B8" -> null
      - pub_key             = ".pubring.gpg" -> null
      - secret_key          = ".secring.gpg" -> null

      - batch {
          - comment = "TF ZTS" -> null
          - email   = "cj@wotw.pro" -> null
          - name    = "CJ Harries" -> null
        }
    }

Plan: 0 to add, 0 to change, 1 to destroy.

Do you really want to destroy all resources?
  Terraform will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes

zts_gpg_keys.test: Destroying... [id=908CB26FEE8FD0AB4482F9367A511F6DDFA194B8]
zts_gpg_keys.test: Destruction complete after 0s

Destroy complete! Resources: 1 destroyed.

$ gpg2 --list-keys 'cj@wotw.pro'
gpg: error reading key: No public key

$ ls ./.*.gpg
zsh: no matches found: ./.*.gpg
```

## Using the Provider

This assumes you're running 64bit Linux. If you're not, figure it out.
```shell-session
cd path/to/terraform-provider-zts
go build
mkdir -p ~/.terraform.d/plugins/linux_amd64/
chmod +x ./terraform-provider-zts
mv ./terraform-provider-zts ~/.terraform.d/plugins/linux_amd64/terraform-provider-zts
```
**NOTE:** This, moreso than other components, is a serious PoC with equal parts PoS. Use at your own risk. The GPG key stuff needs some work.
