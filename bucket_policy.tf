resource "aws_s3_bucket_policy" "bucket" {
  bucket     = module.bucket.id
  policy     = data.aws_iam_policy_document.bucket.json
  depends_on = [module.bucket]
}

/*
   To use an action, a principal must not be denied, but must also have an allow rule
   Denies are strict and explicit, completely limiting actions to only those
   that should have them.
   NOTE: Any actions added for bucket or object should be also added to the relevant DenyUnknown
   statements.
*/
data "aws_iam_policy_document" "bucket" {

  statement {
    sid    = "DenyListBucketContents"
    effect = "Deny"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions = [
      "s3:ListBucket*"
    ]
    resources = [
      module.bucket.arn,
      "${module.bucket.arn}/*"
    ]
    condition {
      test     = "StringNotLike"
      variable = "aws:PrincipalArn"
      values   = local.listers
    }
    condition {
      test     = "StringNotEquals"
      variable = "aws:Service"
      values   = ["access-analyzer.amazonaws.com"]
    }
  }

  statement {
    sid    = "DenyGetBucketActivities"
    effect = "Deny"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions = [
      "s3:GetEncryptionConfiguration",
    ]
    resources = [module.bucket.arn]
    condition {
      test     = "StringNotLike"
      variable = "aws:PrincipalArn"
      values   = sort(distinct(concat(local.readers, local.writers, local.describers)))
    }
    condition {
      test     = "StringNotEquals"
      variable = "aws:Service"
      values   = ["access-analyzer.amazonaws.com"]
    }
  }

  statement {
    sid    = "DenyReadActivities"
    effect = "Deny"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions = [
      "s3:GetObject*",
    ]
    resources = ["${module.bucket.arn}/*"]
    condition {
      test     = "StringNotLike"
      variable = "aws:PrincipalArn"
      values   = local.readers
    }
  }

  statement {
    sid    = "DenyWriteActivities"
    effect = "Deny"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions = [
      "s3:DeleteObject*",
      "s3:PutObject*",
      "s3:ListMultipartUploadParts",
    ]
    resources = ["${module.bucket.arn}/*"]
    condition {
      test     = "StringNotLike"
      variable = "aws:PrincipalArn"
      values   = local.writers
    }
  }

  statement {
    sid    = "DenyBucketMetaActivities"
    effect = "Deny"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions = [
      "s3:GetBucket*",
      "s3:GetLifecycleConfiguration",
      "s3:GetEncryptionConfiguration",
    ]
    resources = [module.bucket.arn]
    condition {
      test     = "StringNotLike"
      variable = "aws:PrincipalArn"
      values   = local.describers
    }
    condition {
      test     = "StringNotEquals"
      variable = "aws:Service"
      values   = ["access-analyzer.amazonaws.com"]
    }
  }

  statement {
    sid    = "DenyAdminActivities"
    effect = "Deny"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions = [
      "s3:DeleteBucket*",
      "s3:GetAccelerateConfiguration",
      "s3:GetAnalyticsConfiguration",
      "s3:GetInventoryConfiguration",
      "s3:GetMetricsConfiguration",
      "s3:GetReplicationConfiguration",
      "s3:PutAccelerateConfiguration",
      "s3:PutAnalyticsConfiguration",
      "s3:PutBucket*",
      "s3:PutEncryptionConfiguration",
      "s3:PutInventoryConfiguration",
      "s3:PutLifecycleConfiguration",
      "s3:PutMetricsConfiguration",
      "s3:PutReplicationConfiguration",
    ]
    resources = [module.bucket.arn]
    condition {
      test     = "StringNotLike"
      variable = "aws:PrincipalArn"
      values   = local.admins
    }
  }

  statement {
    sid    = "DenyUnknownBucketActions"
    effect = "Deny"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    not_actions = [
      "s3:DeleteBucket*",
      "s3:GetAccelerateConfiguration",
      "s3:GetAnalyticsConfiguration",
      "s3:GetBucket*",
      "s3:GetEncryptionConfiguration",
      "s3:GetInventoryConfiguration",
      "s3:GetLifecycleConfiguration",
      "s3:GetMetricsConfiguration",
      "s3:GetReplicationConfiguration",
      "s3:ListBucket*",
      "s3:PutAccelerateConfiguration",
      "s3:PutAnalyticsConfiguration",
      "s3:PutBucket*",
      "s3:PutEncryptionConfiguration",
      "s3:PutInventoryConfiguration",
      "s3:PutLifecycleConfiguration",
      "s3:PutMetricsConfiguration",
      "s3:PutReplicationConfiguration",
    ]
    resources = [module.bucket.arn]
    condition {
      test     = "StringNotLike"
      variable = "aws:PrincipalArn"
      values   = local.admins
    }
  }

  statement {
    sid    = "DenyUnknownObjectActions"
    effect = "Deny"
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    not_actions = [
      "s3:DeleteObject*",
      "s3:GetEncryptionConfiguration",
      "s3:GetObject*",
      "s3:ListBucket*",
      "s3:ListMultipartUploadParts",
      "s3:PutObject*",
    ]
    resources = ["${module.bucket.arn}/*"]
  }

  /*
   This statement makes the bucket usable without further policy
   requirements on the module caller
  */
  statement {
    sid    = "AllowEverythingNotDenied"
    effect = "Allow"
    principals {
      type        = "AWS"
      identifiers = ["*"]
    }
    actions = ["s3:*"]
    resources = [
      module.bucket.arn,
      "${module.bucket.arn}/*",
    ]
    condition {
      test     = "StringLike"
      variable = "aws:PrincipalArn"
      values   = local.all_roles
    }
  }

  statement {
    sid    = "DenyInsecureTransport"
    effect = "Deny"
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    actions = [
      "s3:*",
    ]
    resources = [
      module.bucket.arn,
      "${module.bucket.arn}/*",
    ]
    condition {
      test     = "Bool"
      variable = "aws:SecureTransport"
      values   = ["false"]
    }
  }

  statement {
    sid    = "IfAlgSuppliedMustBeKms"
    effect = "Deny"
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    actions = [
      "s3:PutObject",
    ]
    resources = ["${module.bucket.arn}/*"]
    condition {
      test     = "Null"
      variable = "s3:x-amz-server-side-encryption"
      values   = ["false"]
    }
    condition {
      test     = "StringNotEquals"
      variable = "s3:x-amz-server-side-encryption"
      values   = ["aws:kms"]
    }
  }

  statement {
    sid    = "IfAlgSuppliedKmsKeyMustMatchBucketKey"
    effect = "Deny"
    principals {
      type        = "*"
      identifiers = ["*"]
    }
    actions = [
      "s3:PutObject",
    ]
    resources = ["${module.bucket.arn}/*"]
    condition {
      test     = "Null"
      variable = "s3:x-amz-server-side-encryption"
      values   = ["false"]
    }
    condition {
      test     = "StringNotEquals"
      variable = "s3:x-amz-server-side-encryption-aws-kms-key-id"
      values = [
        module.bucket.kms_key_arn,
        module.bucket.kms_key_id,
      ]
    }
  }

  dynamic "statement" {
    for_each = var.required_tags_with_restricted_values

    content {
      sid    = "DenyPutObjectUnlessTagPresent-${statement.key}"
      effect = "Deny"
      principals {
        type        = "*"
        identifiers = ["*"]
      }
      actions = [
        "s3:PutObject",
      ]
      resources = ["${module.bucket.arn}/*"]
      condition {
        test     = "ForAllValues:StringNotEquals"
        variable = "s3:RequestObjectTagKeys"
        values   = [statement.key]
      }
    }
  }

  dynamic "statement" {
    for_each = var.required_tags_with_restricted_values

    content {
      sid    = "DenyPutObjectUnlessTagMatches-${statement.key}"
      effect = "Deny"
      principals {
        type        = "*"
        identifiers = ["*"]
      }
      actions = [
        "s3:PutObject",
      ]
      resources = ["${module.bucket.arn}/*"]
      condition {
        test     = "StringNotEquals"
        variable = "s3:RequestObjectTag/${statement.key}"
        values   = statement.value
      }
    }
  }

  dynamic "statement" {
    for_each = length(var.restricted_vpce_access) > 0 ? [1] : []

    content {
      sid    = "DenyVPCeAccessUnlessInList"
      effect = "Deny"
      principals {
        type        = "*"
        identifiers = ["*"]
      }
      actions = [
        "s3:*",
      ]
      resources = [
        module.bucket.arn,
        "${module.bucket.arn}/*"
      ]
      condition {
        test     = "StringNotEquals"
        variable = "aws:SourceVpce"
        values   = var.restricted_vpce_access
      }
      condition {
        test     = "StringNotLike"
        variable = "aws:PrincipalArn"
        values   = sort(distinct(concat(local.admins, local.describers)))
      }
      condition {
        test     = "StringNotEquals"
        variable = "aws:Service"
        values   = ["access-analyzer.amazonaws.com"]
      }
    }
  }

  dynamic "statement" {
    for_each = length(var.restricted_ip_access) > 0 ? [1] : []

    content {
      sid    = "DenyIpsAccessUnlessInList"
      effect = "Deny"
      principals {
        type        = "*"
        identifiers = ["*"]
      }
      actions = [
        "s3:*",
      ]
      resources = [
        module.bucket.arn,
        "${module.bucket.arn}/*"
      ]
      condition {
        test     = "NotIpAddress"
        variable = "aws:SourceIp"
        values   = var.restricted_ip_access
      }
      condition {
        test     = "StringNotLike"
        variable = "aws:PrincipalArn"
        values   = sort(distinct(concat(local.admins, local.describers)))
      }
      condition {
        test     = "StringNotEquals"
        variable = "aws:Service"
        values   = ["access-analyzer.amazonaws.com"]
      }
    }
  }
}
