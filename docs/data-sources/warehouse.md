---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "tabular_warehouse Data Source - terraform-provider-tabular"
subcategory: ""
description: |-
  Warehouse data source
---

# tabular_warehouse (Data Source)

Warehouse data source

## Example Usage

```terraform
data "tabular_warehouse" "warehouse" {
  name = "example-warehouse"
}

output "warehouse_id" {
  value = data.tabular_warehouse.warehouse.id
}

output "warehouse_region" {
  value = data.tabular_warehouse.warehouse.region
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) Role Name

### Read-Only

- `id` (String) ID
- `region` (String) Region


