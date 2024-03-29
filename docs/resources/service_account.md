---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tabular_service_account Resource - terraform-provider-tabular"
subcategory: ""
description: |-
  Tabular Service Account
---

# tabular_service_account (Resource)

Tabular Service Account

## Example Usage

```terraform
data "tabular_role" "default" {
  name = "my-role"
}

resource "tabular_service_account" "default" {
  name    = "my-service-account"
  role_id = data.tabular_role.default.id
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Service account name
- `role_id` (String) Role ID

### Read-Only

- `credential_key` (String) Credential ID
- `credential_secret` (String, Sensitive) Credential secret
- `id` (String) Credential ID

## Import

Import is supported using the following syntax:

```shell
# Service Accounts can by imported by the key id. The secret value will not be imported as
# this value is not accessible after service account creation
terraform import tabular_service_account.default "t-"
```
