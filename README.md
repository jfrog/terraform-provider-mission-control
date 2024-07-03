[![Terraform & OpenTofu Acceptance Tests](https://github.com/jfrog/terraform-provider-platform/actions/workflows/acceptance-tests.yml/badge.svg)](https://github.com/jfrog/terraform-provider-platform/actions/workflows/acceptance-tests.yml)

# Terraform Provider for JFrog Mission Control

## Quick Start

Create a new Terraform file with `missioncontrol` resource. Also see [sample.tf](./sample.tf):

### HCL Example

```terraform
# Required for Terraform 1.0 and later
terraform {
  required_providers {
    missioncontrol = {
      source  = "jfrog/mission-control"
      version = "1.0.0"
    }
  }
}

provider "missioncontrol" {
  // supply JFROG_URL and JFROG_ACCESS_TOKEN as env vars
}
```

Initialize Terrform:
```sh
$ terraform init
```

Plan (or Apply):
```sh
$ terraform plan
```

Detailed documentation of the resource and attributes are on [Terraform Registry](https://registry.terraform.io/providers/jfrog/mission-control/latest/docs).

## Versioning

In general, this project follows [semver](https://semver.org/) as closely as we can for tagging releases of the package. We've adopted the following versioning policy:

* We increment the **major version** with any incompatible change to functionality, including changes to the exported Go API surface or behavior of the API.
* We increment the **minor version** with any backwards-compatible changes to functionality.
* We increment the **patch version** with any backwards-compatible bug fixes.

## Contributors

See the [contribution guide](CONTRIBUTIONS.md).

## License

Copyright (c) 2024 JFrog.

Apache 2.0 licensed, see [LICENSE][LICENSE] file.

[LICENSE]: ./LICENSE
