terraform {
  required_version = ">= 0.13.7"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 3.68"
    }
  }
}

locals {
  security_audit_role      = "arn:aws:iam::${data.aws_caller_identity.current.id}:role/RoleSecurityReadOnly"
  current_provisioner_role = data.aws_iam_session_context.current.issuer_arn

  readers = var.read_roles
  writers = var.write_roles

  admins     = sort(distinct(concat(var.admin_roles, [local.current_provisioner_role])))
  describers = sort(distinct(concat(local.admins, [local.security_audit_role], var.metadata_read_roles)))
  listers    = sort(distinct(concat(local.admins, var.list_roles)))
  all_roles  = sort(distinct(concat(local.admins, local.describers, var.read_roles, var.write_roles, var.list_roles)))
}

module "bucket" {
  source             = "hmrc/s3-bucket-core/aws"
  version            = "~> 0.1.4"
  bucket_name        = var.bucket_name
  versioning_enabled = var.versioning_enabled
  data_expiry        = var.data_expiry
  data_sensitivity   = var.data_sensitivity
  force_destroy      = var.force_destroy
  kms_key_policy     = data.aws_iam_policy_document.kms.json
  log_bucket_id      = var.log_bucket_id
  tags               = var.tags
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}
