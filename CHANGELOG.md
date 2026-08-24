## 1.1.1 (August 20, 2026)

SECURITY:

* Bump Go to 1.25.13 and `golang.org/x/net` to v0.58.0 to remediate CVE-2026-39821.
* Remediate CVE-2026-56865 in Go.
* Remediate CVE-2026-56864 in Go.
* Remediate CVE-2026-33818 in Go.
* Remediate CVE-2026-46600 in Go.
* Remediate CVE-2026-56862 in Go.
* Remediate CVE-2026-56859 in Go.
* Remediate CVE-2026-56860 in Go.
* Remediate CVE-2026-56858 in Go.
* Remediate CVE-2026-56853 in Go.
* Remediate CVE-2026-25680 in `golang.org/x/net`.
* Remediate CVE-2026-42506 in `golang.org/x/net`.
* Remediate CVE-2026-42502 in `golang.org/x/net`.
* Remediate CVE-2026-25681 in `golang.org/x/net`.
* Remediate CVE-2026-27136 in `golang.org/x/net`.
* Bump `golang.org/x/crypto` to v0.55.0 to remediate CVE-2026-46595.
* Remediate CVE-2026-42508 in `golang.org/x/crypto`.
* Remediate CVE-2026-39834 in `golang.org/x/crypto`.
* Remediate CVE-2026-39833 in `golang.org/x/crypto`.
* Remediate CVE-2026-39832 in `golang.org/x/crypto`.
* Remediate CVE-2026-39831 in `golang.org/x/crypto`.
* Remediate CVE-2026-39830 in `golang.org/x/crypto`.
* Remediate CVE-2026-39829 in `golang.org/x/crypto`.
* Remediate CVE-2026-46597 in `golang.org/x/crypto`.
* Remediate CVE-2026-39828 in `golang.org/x/crypto`.
* Remediate CVE-2026-39827 in `golang.org/x/crypto`.
* Remediate CVE-2026-39835 in `golang.org/x/crypto`.
* Remediate CVE-2026-46598 in `golang.org/x/crypto`.
* Remediate CVE-2025-47914 in `golang.org/x/crypto`.
* Remediate CVE-2025-58181 in `golang.org/x/crypto`.
* Bump `github.com/cloudflare/circl` to v1.6.5 to remediate CVE-2026-1229.

## 1.1.0 (October 17, 2024). Tested on Artifactory 7.90.14 with Terraform 1.9.8 and OpenTofu 1.8.3

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