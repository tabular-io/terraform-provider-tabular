# Creates a storage profile to host one or more Tabular warehouses
resource "tabular_s3_storage_profile" "default" {
  region         = "us-east-1"
  s3_bucket_name = "s3-bucket-name"
  role_arn       = "IAM Role Arn"
}