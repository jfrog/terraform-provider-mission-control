package missioncontrol_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

func TestAccAccessFederationStar_full(t *testing.T) {
	var skipTest = func() (bool, string) {
		if len(os.Getenv("ARTIFACTORY_URL_2")) > 0 {
			return false, "Env var `ARTIFACTORY_URL_2` is set. Executing test."
		}

		return true, "Env var `ARTIFACTORY_URL_2` is not set. Skipping test."
	}

	if skip, reason := skipTest(); skip {
		t.Skipf(reason)
	}

	_, fqrn, resourceName := testutil.MkNames("test-access-federation", "missioncontrol_access_federation_star")

	temp := `
	resource "missioncontrol_access_federation_star" "{{ .name }}" {
		id = "JPD-1"
		entities = ["USERS", "GROUPS", "PERMISSIONS", "TOKENS"]
		targets = [
			{
				id = "JPD-2"
				url = "http://host.docker.internal:9082/access"
				permission_filters = {
					include_patterns = ["foo", "bar"]
					exclude_patterns = ["fizz", "buzz"]
				}
			},
		]
	}`

	testData := map[string]string{
		"name": resourceName,
	}

	config := util.ExecuteTemplate(resourceName, temp, testData)

	updatedTemp := `
	resource "missioncontrol_access_federation_star" "{{ .name }}" {
		id = "JPD-1"
		entities = ["USERS", "GROUPS", "PERMISSIONS"]
		targets = [
			{
				id = "JPD-2"
				url = "http://host.docker.internal:9082/access"
				permission_filters = {
					include_patterns = ["foo"]
				}
			},
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
					resource.TestCheckResourceAttr(fqrn, "id", "JPD-1"),
					resource.TestCheckResourceAttr(fqrn, "entities.#", "4"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "USERS"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "GROUPS"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "PERMISSIONS"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "TOKENS"),
					resource.TestCheckResourceAttr(fqrn, "targets.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "targets.0.id", "JPD-2"),
					resource.TestCheckResourceAttr(fqrn, "targets.0.url", "http://host.docker.internal:9082/access"),
					resource.TestCheckResourceAttr(fqrn, "targets.0.permission_filters.include_patterns.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "targets.0.permission_filters.include_patterns.*", "foo"),
					resource.TestCheckTypeSetElemAttr(fqrn, "targets.0.permission_filters.include_patterns.*", "bar"),
					resource.TestCheckResourceAttr(fqrn, "targets.0.permission_filters.exclude_patterns.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "targets.0.permission_filters.exclude_patterns.*", "fizz"),
					resource.TestCheckTypeSetElemAttr(fqrn, "targets.0.permission_filters.exclude_patterns.*", "buzz"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "id", "JPD-1"),
					resource.TestCheckResourceAttr(fqrn, "entities.#", "3"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "USERS"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "GROUPS"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "PERMISSIONS"),
					resource.TestCheckResourceAttr(fqrn, "targets.#", "1"),
					resource.TestCheckResourceAttr(fqrn, "targets.0.id", "JPD-2"),
					resource.TestCheckResourceAttr(fqrn, "targets.0.url", "http://host.docker.internal:9082/access"),
					resource.TestCheckResourceAttr(fqrn, "targets.0.permission_filters.include_patterns.#", "1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "targets.0.permission_filters.include_patterns.*", "foo"),
					resource.TestCheckNoResourceAttr(fqrn, "targets.0.permission_filters.exclude_patterns"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
