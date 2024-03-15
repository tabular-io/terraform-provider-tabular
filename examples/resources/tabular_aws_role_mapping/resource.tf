resource "tabular_role" "default" {
  name = "my-role-name"
}

resource "tabular_aws_role_mapping" "default" {
  role_id      = tabular_role.default.id
  aws_role_arn = "arn:aws:iam::123456789012:role/my-iam-role"
}