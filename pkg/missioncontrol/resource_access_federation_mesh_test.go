package missioncontrol_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

// To execute this test, you need setup second Artifactory instance with circle-of-trust.
// Then set them as env vars before running the test
func TestAccAccessFederationMesh_full(t *testing.T) {
	var skipTest = func() (bool, string) {
		if len(os.Getenv("ARTIFACTORY_URL_2")) > 0 {
			return false, "Env var `ARTIFACTORY_URL_2` is set. Executing test."
		}

		return true, "Env var `ARTIFACTORY_URL_2` is not set. Skipping test."
	}

	if skip, reason := skipTest(); skip {
		t.Skipf(reason)
	}

	_, fqrn, resourceName := testutil.MkNames("test-access-federation", "missioncontrol_access_federation_mesh")

	temp := `
	resource "missioncontrol_access_federation_mesh" "{{ .name }}" {
		ids = ["JPD-1", "JPD-2"]
		entities = ["USERS", "GROUPS", "PERMISSIONS", "TOKENS"]
	}`

	testData := map[string]string{
		"name": resourceName,
	}

	config := util.ExecuteTemplate(resourceName, temp, testData)

	updatedTemp := `
	resource "missioncontrol_access_federation_mesh" "{{ .name }}" {
		ids = ["JPD-1", "JPD-2"]
		entities = ["USERS", "GROUPS", "PERMISSIONS"]
	}`
	updatedConfig := util.ExecuteTemplate(resourceName, updatedTemp, testData)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProviders(),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "ids.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "ids.*", "JPD-1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "ids.*", "JPD-2"),
					resource.TestCheckResourceAttr(fqrn, "entities.#", "4"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "USERS"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "GROUPS"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "PERMISSIONS"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "TOKENS"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(fqrn, "ids.#", "2"),
					resource.TestCheckTypeSetElemAttr(fqrn, "ids.*", "JPD-1"),
					resource.TestCheckTypeSetElemAttr(fqrn, "ids.*", "JPD-2"),
					resource.TestCheckResourceAttr(fqrn, "entities.#", "3"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "USERS"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "GROUPS"),
					resource.TestCheckTypeSetElemAttr(fqrn, "entities.*", "PERMISSIONS"),
				),
			},
			{
				ResourceName:      fqrn,
				ImportState:       true,
				ImportStateId:     "JPD-1:JPD-2",
				ImportStateVerify: true,
			},
		},
	})
}
