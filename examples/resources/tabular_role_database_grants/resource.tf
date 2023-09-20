data "tabular_warehouse" "warehouse" {
  name = "funhouse"
}

resource "tabular_role" "example" {
  name = "Example Role 1"
}

resource "tabular_database" "database" {
  warehouse_id = data.tabular_warehouse.warehouse.id
  name         = "other!"
}

resource "tabular_role_database_grants" "grants" {
  role_name    = tabular_role.example.name
  warehouse_id = data.tabular_warehouse.warehouse.id
  database     = tabular_database.database.name
  privileges   = [
    "LIST_TABLES",
  ]
  privileges_with_grant = [
    "FUTURE_DROP_TABLE",
  ]
}