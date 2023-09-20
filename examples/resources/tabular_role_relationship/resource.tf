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