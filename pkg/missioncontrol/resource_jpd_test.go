package missioncontrol_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

// To make tests work runs ./scripts/run-artifactory-2.sh which will export env var `ARTIFACTORY_URL_2`
func TestAccJpd_full(t *testing.T) {
	var skipTest = func() (bool, string) {
		if len(os.Getenv("ARTIFACTORY_URL_2")) > 0 && len(os.Getenv("ARTIFACTORY_JOIN_KEY")) > 0 {
			return false, "Env var `ARTIFACTORY_URL_2` and `ARTIFACTORY_JOIN_KEY` are set. Executing test."
		}

		return true, "Env var `ARTIFACTORY_URL_2` or `ARTIFACTORY_JOIN_KEY` are not set. Skipping test."
	}

	if skip, reason := skipTest(); skip {
		t.Skipf(reason)
	}

	_, fqrn, resourceName := testutil.MkNames("test-jpd", "missioncontrol_jpd")

	temp := `
	resource "missioncontrol_jpd" "{{ .name }}" {
		name = "{{ .name }}"
		url  = "http://host.docker.internal:9082/"
		token  = "{{ .token }}"

		location = {
			city_name = "San Francisco"
			country_code = "US"
			latitude = 37.7749
			longitude = 122.4194
		}

		tags = [
			"prod",
			"dev",
		]
	}`

	// Get the join key from the second Artifactory instance web UI: https://jfrog.com/help/r/jfrog-platform-administration-documentation/view-the-join-key
	// then set the value to env var ARTIFACTORY_2_JOIN_KEY
	testData := map[string]string{
		"name":  resourceName,
		"token": os.Getenv("ARTIFACTORY_JOIN_KEY"),
	}

	config := util.ExecuteTemplate(resourceName, temp, testData)

	updatedTemp := `
	resource "missioncontrol_jpd" "{{ .name }}" {
		name = "{{ .name }}"
		url  = "http://host.docker.internal:9082/"
		token  = "{{ .token }}"

		location = {
			city_name = "New York"
			country_code = "US"
			latitude = 40.7128
			longitude = 74.006
		}

		tags = [
			"dev",
		]
	}`
	updatedConfig := util.ExecuteTemplate(resourceName, updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "url", "http://host.docker.internal:9082/"),
					resource.TestCheckResourceAttr(fqrn, "location.city_name", "San Francisco"),
					resource.TestCheckResourceAttr(fqrn, "location.country_code", "US"),
					resource.TestCheckResourceAttr(fqrn, "location.latitude", "37.7749"),
					resource.TestCheckResourceAttr(fqrn, "location.longitude", "122.4194"),
					resource.TestCheckResourceAttr(fqrn, "tags.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "tags.*", "prod"),
					resource.TestCheckTypeSetElemAttr(fqrn, "tags.*", "dev"),
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttrSet(fqrn, "base_url"),
					resource.TestCheckResourceAttr(fqrn, "status.code", "ONLINE"),
					resource.TestCheckResourceAttrSet(fqrn, "status.message"),
					resource.TestCheckResourceAttr(fqrn, "status.warnings.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "local", "false"),
					resource.TestCheckResourceAttr(fqrn, "services.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "services.0.type", "ARTIFACTORY"),
					resource.TestCheckResourceAttr(fqrn, "services.0.status.code", "ONLINE"),
					resource.TestCheckResourceAttr(fqrn, "licenses.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "is_cold_storage", "false"),
					resource.TestCheckNoResourceAttr(fqrn, "cold_storage_jpd"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "name", testData["name"]),
					resource.TestCheckResourceAttr(fqrn, "url", "http://host.docker.internal:9082/"),
					resource.TestCheckResourceAttr(fqrn, "location.city_name", "New York"),
					resource.TestCheckResourceAttr(fqrn, "location.country_code", "US"),
					resource.TestCheckResourceAttr(fqrn, "location.latitude", "40.7128"),
					resource.TestCheckResourceAttr(fqrn, "location.longitude", "74.006"),
					resource.TestCheckResourceAttr(fqrn, "tags.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "tags.*", "dev"),
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttrSet(fqrn, "base_url"),
					resource.TestCheckResourceAttr(fqrn, "status.code", "ONLINE"),
					resource.TestCheckResourceAttrSet(fqrn, "status.message"),
					resource.TestCheckResourceAttr(fqrn, "status.warnings.#", "0"),
					resource.TestCheckResourceAttr(fqrn, "local", "false"),
					resource.TestCheckResourceAttr(fqrn, "services.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "services.0.type", "ARTIFACTORY"),
					resource.TestCheckResourceAttr(fqrn, "services.0.status.code", "ONLINE"),
					resource.TestCheckResourceAttr(fqrn, "licenses.#", "1"),
					resource.TestCheckResourceAttrSet(fqrn, "licenses.0.type"),
					resource.TestCheckResourceAttrSet(fqrn, "licenses.0.expired"),
					resource.TestCheckResourceAttrSet(fqrn, "licenses.0.license_hash"),
					resource.TestCheckResourceAttrSet(fqrn, "licenses.0.licensed_to"),
					resource.TestCheckResourceAttrSet(fqrn, "licenses.0.valid_through"),
					resource.TestCheckResourceAttr(fqrn, "is_cold_storage", "false"),
					resource.TestCheckNoResourceAttr(fqrn, "cold_storage_jpd"),
				),
			},
			{
				ResourceName:            fqrn,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"token", "username", "password"},
			},
		},
	})
}
