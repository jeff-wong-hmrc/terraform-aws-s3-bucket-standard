output "bucket_name" {
  value = module.s3_example.id
}

output "bucket_kms_key_arn" {
  value = module.s3_example.kms_key_arn
}

output "bucket_kms_key_id" {
  value = module.s3_example.kms_key_id
}

output "additional_kms_key_arn" {
  value = aws_kms_key.additional_key.arn
}

output "object_lock" {
  value = local.object_lock
}
