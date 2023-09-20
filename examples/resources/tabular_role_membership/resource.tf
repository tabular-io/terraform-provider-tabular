resource "tabular_role" "example" {
  name = "Example Role"
}

resource "tabular_role_membership" "example_members" {
  role_name     = tabular_role.example.name
  admin_members = ["role_admin@tabular.io"]
  members       = ["user@tabular.io"]
}