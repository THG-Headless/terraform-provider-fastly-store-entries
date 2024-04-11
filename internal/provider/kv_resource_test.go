// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func randomString(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	ret := make([]byte, n)
	for i := 0; i < n; i++ {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		ret[i] = letters[num.Int64()]
	}

	return string(ret)
}

var TEST_KV_ID = os.Getenv("TEST_KV_ID")
var TEST_KV_KEY = randomString(10)

func TestAccKVResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccKVResourceConfig(TEST_KV_ID, TEST_KV_KEY, "testValue"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fastlystoreitems_kv.tester", "store_id", TEST_KV_ID),
					resource.TestCheckResourceAttr("fastlystoreitems_kv.tester", "key", TEST_KV_KEY),
					resource.TestCheckResourceAttr("fastlystoreitems_kv.tester", "value", "testValue"),
				),
			},
			{
				Config: testAccKVResourceConfig(TEST_KV_ID, TEST_KV_KEY, "testValue"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("fastlystoreitems_kv.tester", "store_id", TEST_KV_ID),
					resource.TestCheckResourceAttr("fastlystoreitems_kv.tester", "key", TEST_KV_KEY),
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
