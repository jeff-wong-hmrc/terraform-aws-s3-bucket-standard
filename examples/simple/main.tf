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
  read_roles  = [aws_iam_role.read.arn, data.aws_iam_session_context.current.issuer_arn]
  write_roles = [aws_iam_role.write.arn, data.aws_iam_session_context.current.issuer_arn]
  list_roles  = [aws_iam_role.list.arn, data.aws_iam_session_context.current.issuer_arn]
  admin_roles = [aws_iam_role.admin.arn, data.aws_iam_session_context.current.issuer_arn]
  metadata_read_roles = [aws_iam_role.metadata.arn]

  data_expiry      = "90-days"
  data_sensitivity = "low"
  force_destroy    = true
  tags             = var.tags
  log_bucket_id    = aws_s3_bucket.access_logs.id
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}

data "aws_iam_policy_document" "instance-assume-role-policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = [data.aws_iam_session_context.current.issuer_arn]
    }
  }
}

resource "aws_iam_role" "read" {
  name               = "${var.test_name}-read-role"
  assume_role_policy = data.aws_iam_policy_document.instance-assume-role-policy.json
}

resource "aws_iam_role" "write" {
  name               = "${var.test_name}-write-role"
  assume_role_policy = data.aws_iam_policy_document.instance-assume-role-policy.json

}

resource "aws_iam_role" "list" {
  name               = "${var.test_name}-list-role"
  assume_role_policy = data.aws_iam_policy_document.instance-assume-role-policy.json
}

resource "aws_iam_role" "admin" {
  name               = "${var.test_name}-admin-role"
  assume_role_policy = data.aws_iam_policy_document.instance-assume-role-policy.json
}

resource "aws_iam_role" "metadata" {
  name               = "${var.test_name}-metadata-role"
  assume_role_policy = data.aws_iam_policy_document.instance-assume-role-policy.json
}
