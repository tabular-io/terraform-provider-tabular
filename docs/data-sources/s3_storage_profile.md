---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tabular_s3_storage_profile Data Source - terraform-provider-tabular"
subcategory: ""
description: |-
  S3StorageProfile data source
---

# tabular_s3_storage_profile (Data Source)

S3StorageProfile data source

## Example Usage

```terraform
data "tabular_s3_storage_profile" "test" {
  name = "my-s3-bucket-name"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Storage Profile bucket name

### Read-Only

- `account_id` (String) Storage Profile AWS Account ID
- `external_id` (String) External ID
- `id` (String) S3StorageProfile ID
- `organization_id` (String) Tabular Organization ID
- `region` (String) Storage Profile region
- `role_arn` (String) Storage Profile AWS Role Arn
