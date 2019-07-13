provider "zts" {
  etcd_host = "http://127.0.0.1:2379/"
}

resource "zts_secrets" "test" {
  etcd_key = "/test-client/secrets.json"
}

resource "zts_gpg_keys" "test" {
  directory = "."
  batch {
    name    = "CJ Harries"
    email   = "cj@wotw.pro"
    comment = "TF Zero Trust Secrets"
  }
}
