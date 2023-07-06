terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "5.4.0"
    }
  }
}

provider "aws" {
  region = "eu-west-2"
}

locals {
  provisioner_role = data.aws_iam_session_context.current.issuer_arn
  object_lock      = false
}

module "s3_example" {
  source = "../../"
  #  source      = "hmrc/s3-bucket-standard/aws"
  bucket_name = "${var.test_name}-bucket"

  read_roles  = [local.provisioner_role]
  write_roles = [local.provisioner_role]

  data_expiry      = "1-day"
  data_sensitivity = "low"
  force_destroy    = true
  log_bucket_id    = aws_s3_bucket.access_logs.id
  object_lock      = local.object_lock
  object_lock_mode = "GOVERNANCE"
}

resource "aws_kms_key" "additional_key" {
  description         = "Key that we should not be able to use when putting an object"
  enable_key_rotation = true
  policy              = data.aws_iam_policy_document.additional.json
}

data "aws_iam_policy_document" "additional" {
  statement {
    effect  = "Allow"
    actions = ["kms:*"]
    principals {
      identifiers = [local.provisioner_role]
      type        = "AWS"
    }
    resources = ["*"]
  }
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}