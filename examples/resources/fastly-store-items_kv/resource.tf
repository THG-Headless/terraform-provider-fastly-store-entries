resource "fastly-store-items_kv" "item" {
  store_id = "{{STOREID}}"
  key      = "foo"
  value    = "bar"
}
