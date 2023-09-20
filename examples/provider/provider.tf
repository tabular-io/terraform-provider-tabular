terraform {
  required_providers {
    tabular = {
      source = "tabular-io/tabular"
    }
  }
}

provider "tabular" {
  organization_id = var.organization_id
}