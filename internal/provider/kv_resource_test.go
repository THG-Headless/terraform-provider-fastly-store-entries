// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

var TEST_KV_ID = os.Getenv("TEST_KV_ID")

func TestAccKVResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKVResourceConfig(TEST_KV_ID, "testKey", "testValue"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fastlystoreitems_kv.tester", "store_id", TEST_KV_ID),
					resource.TestCheckResourceAttr("fastlystoreitems_kv.tester", "key", "testKey"),
					resource.TestCheckResourceAttr("fastlystoreitems_kv.tester", "value", "testValue"),
				),
			},
			{
				Config: testAccKVResourceConfig(TEST_KV_ID, "testKey", "testValue"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fastlystoreitems_kv.tester", "store_id", TEST_KV_ID),
					resource.TestCheckResourceAttr("fastlystoreitems_kv.tester", "key", "testKey"),
					resource.TestCheckResourceAttr("fastlystoreitems_kv.tester", "value", "testValue"),
				),
			},
		},
	})
}

func testAccKVResourceConfig(storeId string, key string, value string) string {
	return fmt.Sprintf(`
resource "fastlystoreitems_kv" "tester" {
	store_id = %q
	key      = %q
	value    = %q
}
`, storeId, key, value)
}
