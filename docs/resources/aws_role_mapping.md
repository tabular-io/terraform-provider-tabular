---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tabular_aws_role_mapping Resource - terraform-provider-tabular"
subcategory: ""
description: |-
  Tabular AWS Role Mapping
---

# tabular_aws_role_mapping (Resource)

Tabular AWS Role Mapping

## Example Usage

```terraform
resource "tabular_role" "default" {
  name = "my-role-name"
}

resource "tabular_aws_role_mapping" "default" {
  role_id      = tabular_role.default.id
  aws_role_arn = "arn:aws:iam::123456789012:role/my-iam-role"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `aws_role_arn` (String) AWS IAM Role ARN
- `role_id` (String) Role ID

### Read-Only

- `id` (String) Credential ID
- `name` (String)