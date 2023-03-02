terraform {
  required_providers {
    tabular = {
      source = "tabular-io/tabular"
    }
  }
}

provider "tabular" {}

resource "tabular_role" "example" {
  name          = "Example Role 1"
  force_destroy = true
}

resource "tabular_role" "example2" {
  name          = "Example Role 2"
  force_destroy = true
}

resource "tabular_role" "example3" {
  name          = "Example Role 3"
  force_destroy = true
}

resource "tabular_role_relationship" "inheritance" {
  parent_role_name = tabular_role.example.name
  child_role_name  = tabular_role.example2.name
}

resource "tabular_role_database_grants" "grants" {
  role_name    = tabular_role.example.name
  warehouse_id = "fb0723be-72e7-414c-b060-0a4e3c6d8cdc"
  database     = "dirt"
  privileges = [
    "CREATE_TABLE",
    "LIST_TABLES",
    "MODIFY_DATABASE"
  ]
  privileges_with_grant = [
    "FUTURE_DROP_TABLE",
    "FUTURE_SELECT",
    "FUTURE_UPDATE"
  ]
}