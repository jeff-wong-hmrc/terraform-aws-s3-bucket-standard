terraform {
  required_version = ">= 0.13.7"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 4.9"
    }
  }
}

locals {
  current_provisioner_role = data.aws_iam_session_context.current.issuer_arn

  security_audit_role  = var.allow_security_team_metadata_audit ? ["arn:aws:iam::${data.aws_caller_identity.current.id}:role/RoleSecurityReadOnly"] : []
  guardduty_audit_role = var.allow_guardduty_metadata_audit ? ["arn:aws:iam::${data.aws_caller_identity.current.id}:role/aws-service-role/guardduty.amazonaws.com/AWSServiceRoleForAmazonGuardDuty"] : []

  readers = var.read_roles
  writers = var.write_roles

  default_services       = ["access-analyzer.amazonaws.com"]
  read_services          = sort(distinct(concat(local.default_services, var.read_services)))
  list_services          = sort(distinct(concat(local.default_services, var.list_services)))
  metadata_read_services = sort(distinct(concat(local.default_services, var.metadata_read_services)))
  write_services         = var.write_services


  admins     = sort(distinct(concat(var.admin_roles, [local.current_provisioner_role])))
  describers = sort(distinct(concat(local.admins, local.security_audit_role, local.guardduty_audit_role, var.metadata_read_roles)))
  listers    = sort(distinct(concat(local.admins, var.list_roles)))
  all_roles  = sort(distinct(concat(local.admins, local.describers, var.read_roles, var.write_roles, var.list_roles)))
}

module "bucket" {
  source                     = "github.com/hmrc/terraform-aws-s3-bucket-core?ref=3.0.1"
  bucket_name                = var.bucket_name
  versioning_enabled         = var.versioning_enabled
  data_expiry                = var.data_expiry
  data_sensitivity           = var.data_sensitivity
  force_destroy              = var.force_destroy
  kms_key_policy             = data.aws_iam_policy_document.kms.json
  log_bucket_id              = var.log_bucket_id
  tags                       = var.tags
  transition_to_glacier_days = var.transition_to_glacier_days
  object_lock                = var.object_lock
  object_lock_mode           = var.object_lock_mode
}

data "aws_caller_identity" "current" {}

data "aws_iam_session_context" "current" {
  arn = data.aws_caller_identity.current.arn
}
