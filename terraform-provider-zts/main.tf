provider "zts" {
  control_server = "localhost:8080"
}

resource "zts_secrets" "test" {
  etcd_host = "http://127.0.0.1:2379/"
  etcd_key  = "/test-client/secrets.json"
}
