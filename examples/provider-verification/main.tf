terraform {
  required_providers {
    fastlystoreitems = {
      source = "fastly-store-items"
    }
    fastly = {
      source = "fastly/fastly"
    }
  }
}

provider "fastlystoreitems" {
  api_key = "{{APIKEY}}"
}

resource "fastlystoreitems_kv" "item_1" {
  store_id = "{{STOREID}}"
  key      = "foo"
  value    = "bar"
}
