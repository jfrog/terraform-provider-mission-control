package missioncontrol_test

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/jfrog/terraform-provider-shared/testutil"
	"github.com/jfrog/terraform-provider-shared/util"
)

// To execute this test, you need a signed license bucket URL and key from MyJFrog
// (Note: the signed URL will expired and require fetching a new one)
// Then set them as env vars before running the test
func TestAccLicenseBucket_url(t *testing.T) {
	jfrogLicenseBucketURL := os.Getenv("JFROG_LICENSE_BUCKET_URL")
	if jfrogLicenseBucketURL == "" {
		t.Skipf("env var JFROG_LICENSE_BUCKET_URL not set")
	}

	jfrogLicenseBucketKey := os.Getenv("JFROG_LICENSE_BUCKET_KEY")
	if jfrogLicenseBucketKey == "" {
		t.Skipf("env var JFROG_LICENSE_BUCKET_KEY not set")
	}

	_, fqrn, resourceName := testutil.MkNames("test-license-bucket", "missioncontrol_license_bucket")

	temp := `
	resource "missioncontrol_license_bucket" "{{ .name }}" {
		name = "{{ .name }}"
		url  = "{{ .url }}"
		key  = "{{ .key }}"
	}`

	testData := map[string]string{
		"name": resourceName,
		"url":  jfrogLicenseBucketURL,
		"key":  jfrogLicenseBucketKey,
	}

	config := util.ExecuteTemplate(resourceName, temp, testData)

	updatedTemp := `
	resource "missioncontrol_license_bucket" "{{ .name }}" {
		name = "{{ .name }}-2"
		url  = "{{ .url }}"
		key  = "{{ .key }}"
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
					resource.TestCheckResourceAttr(fqrn, "url", testData["url"]),
					resource.TestCheckResourceAttr(fqrn, "key", testData["key"]),
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "subject", "JFROG TEST"),
					resource.TestCheckResourceAttr(fqrn, "valid_date", "2025-01-04T00:00:00.000Z"),
					resource.TestCheckResourceAttr(fqrn, "issued_date", "2023-12-23T08:46:48.000Z"),
					resource.TestCheckResourceAttr(fqrn, "product_name", "Multiproducts"),
					resource.TestCheckResourceAttr(fqrn, "product_id", "7"),
					resource.TestCheckResourceAttr(fqrn, "license_type", "ENTERPRISE_PLUS_TRIAL"),
					resource.TestCheckResourceAttr(fqrn, "quantity", "5"),
					resource.TestCheckResourceAttr(fqrn, "used", "0"),
				),
			},
			{
				Config: updatedConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(fqrn, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

// To execute this test, you need the encrypted file (download from the signed license bucket URL)
// and key from MyJFrog (Note: the signed URL will expired and require fetching a new one)
// Then set the file path and key as env vars before running the test
func TestAccLicenseBucket_file(t *testing.T) {
	jfrogLicenseBucketFile := os.Getenv("JFROG_LICENSE_BUCKET_FILE")
	if jfrogLicenseBucketFile == "" {
		t.Skipf("env var JFROG_LICENSE_BUCKET_FILE not set")
	}

	jfrogLicenseBucketKey := os.Getenv("JFROG_LICENSE_BUCKET_KEY")
	if jfrogLicenseBucketKey == "" {
		t.Skipf("env var JFROG_LICENSE_BUCKET_KEY not set")
	}

	_, fqrn, resourceName := testutil.MkNames("test-license-bucket", "missioncontrol_license_bucket")

	temp := `
	resource "missioncontrol_license_bucket" "{{ .name }}" {
		name = "{{ .name }}"
		file = "{{ .file }}"
		key  = "{{ .key }}"
	}`

	testData := map[string]string{
		"name": resourceName,
		"file": jfrogLicenseBucketFile,
		"key":  jfrogLicenseBucketKey,
	}

	config := util.ExecuteTemplate(resourceName, temp, testData)

	updatedTemp := `
	resource "missioncontrol_license_bucket" "{{ .name }}" {
		name = "{{ .name }}-2"
		file = "{{ .file }}"
		key  = "{{ .key }}"
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
					resource.TestCheckNoResourceAttr(fqrn, "url"),
					resource.TestCheckResourceAttr(fqrn, "key", testData["key"]),
					resource.TestCheckResourceAttrSet(fqrn, "id"),
					resource.TestCheckResourceAttr(fqrn, "subject", "JFROG TEST"),
					resource.TestCheckResourceAttr(fqrn, "valid_date", "2025-01-04T00:00:00.000Z"),
					resource.TestCheckResourceAttr(fqrn, "issued_date", "2023-12-23T08:46:48.000Z"),
					resource.TestCheckResourceAttr(fqrn, "product_name", "Multiproducts"),
					resource.TestCheckResourceAttr(fqrn, "product_id", "7"),
					resource.TestCheckResourceAttr(fqrn, "license_type", "ENTERPRISE_PLUS_TRIAL"),
					resource.TestCheckResourceAttr(fqrn, "quantity", "5"),
					resource.TestCheckResourceAttr(fqrn, "used", "0"),
				),
			},
			{
				Config: updatedConfig,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(fqrn, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}
