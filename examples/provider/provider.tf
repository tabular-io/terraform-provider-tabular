terraform {
  required_providers {
    tabular = {
      source = "tabular-io/tabular"
    }
  }
}

provider "tabular" {}

resource "tabular_role" "example" {
  name = "Example Role 1"
}

resource "tabular_role" "example2" {
  name = "Example Role 2"
}

resource "tabular_role_relationship" "inheritance" {
  parent_role_name = tabular_role.example.name
  child_role_name  = tabular_role.example2.name
}

data "tabular_warehouse" "warehouse" {
  name = "funhouse"
}

resource "tabular_role_database_grants" "grants" {
  role_name    = tabular_role.example.name
  warehouse_id = data.tabular_warehouse.warehouse.id
  database     = "dirt"
  privileges = [
    "LIST_TABLES",
  ]
  privileges_with_grant = [
    "FUTURE_DROP_TABLE",
  ]
}