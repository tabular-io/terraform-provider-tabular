data "tabular_role" "default" {
  name = "my-role"
}

resource "tabular_service_account" "default" {
  name    = "my-service-account"
  role_id = data.tabular_role.default.id
}