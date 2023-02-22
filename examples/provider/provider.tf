terraform {
  required_providers {
    tabular = {
      source = "tabular-io/tabular"
    }
  }
}

provider "tabular" {
  organization_id = "78a1a843-0fe6-4cb1-b9e3-4a5c04f1e48b"
}

data "tabular_role" "data_ex" {
  id = "9990834b-41aa-4517-8f8d-0f31396b2f32"
}

resource "tabular_role" "example" {
  name = "Example Role"
}

output "role_id" {
  value = tabular_role.example.name
}

output "data_role_name" {
  value = data.tabular_role.data_ex.name
}