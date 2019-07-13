provider "zts" {}

resource "zts_server" "my-server" {
  address = "1.2.3.4"
}
