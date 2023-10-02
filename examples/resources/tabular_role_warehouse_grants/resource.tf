data "tabular_role" "terraform" {
  name = "Terraform"
}

data "tabular_warehouse" "terraform" {
  name = "example-warehouse"
}

resource "tabular_role_warehouse_grants" "test" {
  role_id      = data.tabular_role.terraform.id
  warehouse_id = data.tabular_warehouse.terraform.id
  privileges   = [
    "MODIFY_WAREHOUSE",
    "FUTURE_LIST_TABLES"
  ]
  privileges_with_grant = [
    "FUTURE_MODIFY_DATABASE"
  ]
}