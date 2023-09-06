resource "tabular_s3_storage_profile" "default" {
  region         = "us-east-1"
  s3_bucket_name = "s3-bucket-name"
  role_arn       = "IAM Role Arn"
}

resource "tabular_warehouse" "default" {
  name            = "tabular-warehouse-name"
  storage_profile = tabular_s3_storage_profile.default.id
}