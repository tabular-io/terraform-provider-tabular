terraform {
  required_providers {
    tabular = {
      source = "github.com/tabular-io/tabular"
    }
  }
}

provider "tabular" {
  endpoint = "https://api.dev.tabular.io"
}

data "tabular_role" "example" {
  id = "a4ef05fa-5a4e-49b4-9403-2a74044f1bf0"
}

data "tabular_role" "example2" {
  id = "22a3a50d-42fe-4c52-9827-3522a4cc6daf"
}

output "role_id" {
  value = data.tabular_role.example.display_name
}

output "role_id2" {
  value = data.tabular_role.example2.display_name
}