## 1.1.0 (October 17, 2024)

IMPROVEMENTS:

* provider: Add `tfc_credential_tag_name` configuration attribute to support use of different/[multiple Workload Identity Token in Terraform Cloud Platform](https://developer.hashicorp.com/terraform/cloud-docs/workspaces/dynamic-provider-credentials/manual-generation#generating-multiple-tokens). Issue: [#68](https://github.com/jfrog/terraform-provider-shared/issues/68) PR: [#24](https://github.com/jfrog/terraform-provider-mission-control/pull/24)

## 1.0.2 (July 16, 2024). Tested on Artifactory 7.84.17 with Terraform 1.9.2 and OpenTofu 1.7.3

IMPROVEMENTS:

* resource/missioncontrol_jpd: Fix configuration sample and import in documentation. PR: [#11](https://github.com/jfrog/terraform-provider-mission-control/pull/11)

## 1.0.1 (July 16, 2024). Tested on Artifactory 7.84.17 with Terraform 1.9.2 and OpenTofu 1.7.3

IMPROVEMENTS:

* resource/missioncontrol_access_federation_mesh, resource/missioncontrol_access_federation_star, resource/missioncontrol_jpd: Fix documentation formatting. PR: [#10](https://github.com/jfrog/terraform-provider-mission-control/pull/10)

## 1.0.0 (July 16, 2024). Tested on Artifactory 7.84.17 with Terraform 1.9.2 and OpenTofu 1.7.3

FEATURES:

* **New Resource:** `missioncontrol_license_bucket` PR: [#2](https://github.com/jfrog/terraform-provider-mission-control/pull/2)
* **New Resource:** `missioncontrol_jpd` PR: [#3](https://github.com/jfrog/terraform-provider-mission-control/pull/3)
* **New Resource:** `missioncontrol_access_federation_star` and `missioncontrol_access_federation_mesh` PR: [#8](https://github.com/jfrog/terraform-provider-mission-control/pull/8)