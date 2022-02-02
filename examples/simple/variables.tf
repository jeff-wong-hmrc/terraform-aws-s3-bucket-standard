
variable "test_name" {
  description = "short name to ensure unique resources are created during test runs"
  type        = string
}

variable "tags" {
  type        = map(string)
  description = "Additional tags to apply to the bucket"
  default     = {}
}