terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "3.68.0"
    }
  }
}

provider "aws" {
  region = "eu-west-2"
}

module "s3_example" {
  source = "../../"
  #  source      = "hmrc/s3-bucket-standard/aws"
  bucket_name = "${var.test_name}-bucket"

  data_expiry       = "1-day"
  data_sensitivity  = "low"
  force_destroy     = true
  log_bucket_id     = aws_s3_bucket.access_logs.id
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}