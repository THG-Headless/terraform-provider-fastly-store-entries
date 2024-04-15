terraform {
  required_providers {
    fastlystoreitems = {
      version = "0.0.2"
      source  = "thg-headless/fastly-store-items"
    }
    fastly = {
      source = "fastly/fastly"
    }
  }
}

provider "fastly-store-items" {
  api_key = "{{APIKEY}}"
}

resource "fastly-store-items_kv" "item_1" {
  store_id = "{{STOREID}}"
  key      = "foo"
  value    = "bar"
}
