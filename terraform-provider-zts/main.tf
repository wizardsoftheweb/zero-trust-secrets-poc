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

resource "zts_gpg_keys" "test" {
  directory = "."
  batch {
    name    = "CJ Harries"
    email   = "cj@wotw.pro"
    comment = "TF ZTS"
  }
}

output "gpg_key_id" {
  value = "${zts_gpg_keys.test.gpg_key_id}"
}
