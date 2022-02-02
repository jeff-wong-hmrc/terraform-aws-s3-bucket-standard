# S3 bucket standard module

This module creates an AWS S3 bucket, KMS key and associated resources while enforcing the PlatSec S3 bucket policy.
It uses the [core bucket module](https://github.com/hmrc/terraform-aws-s3-bucket-core) which documents the parts of the
policy that it enforces.

This module accepts IAM roles and configures access in the bucket and KMS key resource policies to `Allow` access as
well as `Deny` everyone else. When using this module no additional S3 or KMS permissions are required for access to the
bucket. You will need to ensure there are no `Deny` rules anywhere else in the policy chain
([Policies and permissions in IAM](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies.html).)

This module creates a simple general purpose bucket but does not support AWS service account access such as
`vpc-flow-logs.amazonaws.com` or more complex configuration such as replication.

## Policy Roles

### read_roles

The role or roles which will allow access for reading (and decrypting) objects from the bucket.  
**admin_roles and metadata_read_roles should __never__ be given read access**

### write_roles

The role or roles which will allow access for writing (and encrypting) objects to the bucket.   
**This role can also delete objects.**  
**admin_roles and metadata_read_roles should __never__ be given write access**

### list_roles

The role or roles which will allow access to list the contents of the bucket.  
**admin_roles are also able to list the contents of the bucket** - this is required to successfully manage the bucket with
terraform.

### admin_roles

The role or roles that will administer the bucket and KMS configuration/policy etc.  
__The `admin_roles` are typically your Terraform/Administrator/Jenkins roles__

The administrator roles do not have access to read or write/delete by default **but they can list all objects**.  
They exist to perform any configuration on the bucket/KMS or policies.

If you use more than one role to run terraform ensure they are all present in the admin_roles variable.

The current role that is used when applying this module will be added to the list of admins if it is not already
present.

### metadata_read_roles

The role or roles that can read bucket and KMS key metadata.  
These roles exist so that the bucket and associated resources can be audited to ensure that they comply with the PlatSec
S3 bucket policy.  
**The role `role/RoleSecurityReadOnly` in the AWS account of the bucket is automatically added to the
metadata_read_roles.**  
The security readonly role is not created/managed by this module.

## Policy enforcement

#### Platform Security meta-data access

**AWS role RoleSecurityReadOnly has meta-data access to all S3 buckets.** This is for auditing and alerting purposes

#### Logging

Enforced in [core bucket module](https://github.com/hmrc/terraform-aws-s3-bucket-core)

#### Bucket policy

**At all times PlatSec must have access to all AWS S3 metadata (not the data).** See Platform Security meta-data access
above.

**Secure transport is required.**

**Public access is blocked.** Enforced in [core bucket module](https://github.com/hmrc/terraform-aws-s3-bucket-core)

Additional network level restriction to a fixed known range of IPs or VPCs and locking down access to a set of known IAM
roles **should** be implemented. This module supports limiting access to specific IP addresses or VPCs.  
Buckets containing PII **must** implement IP or VPC restrictions.  
**admin_roles and metadata_read_roles are excluded from the IP/VPC restrictions - DO NOT GIVE THESE ROLES READ OR WRITE
ACCESS TO DATA** 

#### Tagging, Versioning, Life cycle policies, Encryption at rest, Access control list (ACL)

See [core bucket module](https://github.com/hmrc/terraform-aws-s3-bucket-core)

## Additional behaviours

#### Ensure the bucket KMS key is used to encrypt objects

This module ensures that the only encryption key that can be use is the KMS key created by the
[core bucket module](https://github.com/hmrc/terraform-aws-s3-bucket-core).  
To ensure continued access to data, when putting an object either of the following is required
1. do not specify either `x-amz-server-side-encryption` or `x-amz-server-side-encryption-aws-kms-key-id` so that the 
bucket server_side_encryption_configuration takes effect
2. set `x-amz-server-side-encryption` to `aws:kms` and set `x-amz-server-side-encryption-aws-kms-key-id` to the ARN or
the ID of the bucket KMS key

## Tests

### How to use / run tests
In order to integrate with AWS, we need to provide the relevant credentials.
This is done through passing AWS environment variables to the docker container and then, depending on your AWS config set up,
you will need to run the following command in order to pass the credentials through to terraform:

AWS Vault

```aws-vault exec <role> -- make test ```

AWS Profile

``` aws-profile -p <role> make test ```
