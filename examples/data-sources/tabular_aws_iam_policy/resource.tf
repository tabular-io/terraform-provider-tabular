data "tabular_aws_iam_policy" "default" {
  bucket = "my-bucket-name"
}

# Create AWS IAM role with the read-write policy
resource "aws_iam_role" "read_write" {
  name = "my-role"
  assume_role_policy = data.tabular_aws_iam_policy.default.assume_role_policy
  
  inline_policy {
    name = "tabular-access"
    policy = data.tabular_aws_iam_policy.default.iam_read_write_policy
  }
}

# Create AWS IAM role with the read-only policy
resource "aws_iam_role" "read_only" {
  name = "my-role"
  assume_role_policy = data.tabular_aws_iam_policy.default.assume_role_policy
  
  inline_policy {
    name = "tabular-access"
    policy = data.tabular_aws_iam_policy.default.iam_read_only_policy
  }
}
