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
  child_role_name  = tabular_role.example.name
}