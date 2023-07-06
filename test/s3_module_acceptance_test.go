package test

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
	"log"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/gruntwork-io/terratest/modules/random"
	"github.com/gruntwork-io/terratest/modules/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const region = "eu-west-2"

func TestReadRole(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testKey := "testS3KeyName"
	body := "banana"
	terraformOptions := copyTerraformAndReturnOptions(t, "examples/simple", map[string]interface{}{})
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	bucketName := terraform.Output(t, terraformOptions, "bucket_name")

	versionId := CreateTestObject(t, ctx, bucketName, testKey)

	readS3Client := S3ClientFromOutputArn(t, ctx, terraformOptions, "read_role_arn")
	_, err := readS3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    &testKey,
	})
	assert.NoErrorf(t, err, "Could not fetch object for read role")

	_, err = readS3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(bucketName)})
	assert.Error(t, err)

	_, err = readS3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:     aws.String(bucketName),
		Key:        aws.String(testKey),
		Body:       strings.NewReader(body),
		ContentMD5: aws.String(md5EncodeBody(body)),
	})
	require.Error(t, err)

	_, err = readS3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		VersionId: versionId,
		Bucket:    aws.String(bucketName),
		Key:       &testKey,
	})
	require.Error(t, err)

	_, err = readS3Client.ListBucketInventoryConfigurations(ctx, &s3.ListBucketInventoryConfigurationsInput{
		Bucket: aws.String(bucketName),
	})
	assert.Error(t, err)

	_, err = readS3Client.GetBucketAccelerateConfiguration(ctx, &s3.GetBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucketName),
	})
	assert.Error(t, err)
}

func TestWriteRole(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testKey := "testS3KeyName"
	body := "banana"
	terraformOptions := copyTerraformAndReturnOptions(t, "examples/simple", map[string]interface{}{})
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	bucketName := terraform.Output(t, terraformOptions, "bucket_name")

	//test write role
	writeS3Client := S3ClientFromOutputArn(t, ctx, terraformOptions, "write_role_arn")
	uploadResponse, err := writeS3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:     aws.String(bucketName),
		Key:        aws.String(testKey),
		Body:       strings.NewReader(body),
		ContentMD5: aws.String(md5EncodeBody(body)),
	})
	require.NoError(t, err)

	_, err = writeS3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    &testKey,
	})
	assert.Error(t, err)

	_, err = writeS3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(bucketName)})
	assert.Error(t, err)

	_, err = writeS3Client.ListBucketInventoryConfigurations(ctx, &s3.ListBucketInventoryConfigurationsInput{
		Bucket: aws.String(bucketName),
	})
	assert.Error(t, err)

	_, err = writeS3Client.GetBucketAccelerateConfiguration(ctx, &s3.GetBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucketName),
	})
	assert.Error(t, err)

	//test delete on write role
	_, err = writeS3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		VersionId: uploadResponse.VersionId,
		Bucket:    aws.String(bucketName),
		Key:       &testKey,
	})
	fmt.Println(err)
	require.NoError(t, err)
}

func TestListRole(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testKey := "testS3KeyName"
	body := "banana"
	terraformOptions := copyTerraformAndReturnOptions(t, "examples/simple", map[string]interface{}{})
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	bucketName := terraform.Output(t, terraformOptions, "bucket_name")

	versionId := CreateTestObject(t, ctx, bucketName, testKey)
	defer DeleteTestObject(t, ctx, bucketName, testKey, versionId)

	listS3Client := S3ClientFromOutputArn(t, ctx, terraformOptions, "list_role_arn")
	ListResponse, err := listS3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(bucketName)})
	assert.NoError(t, err)
	if err == nil {
		assert.Len(t, ListResponse.Contents, 1)
	}

	_, err = listS3Client.GetBucketAccelerateConfiguration(ctx, &s3.GetBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucketName),
	})
	assert.Error(t, err)

	_, err = listS3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:     aws.String(bucketName),
		Key:        aws.String(testKey),
		Body:       strings.NewReader(body),
		ContentMD5: aws.String(md5EncodeBody(body)),
	})
	require.Error(t, err)

	_, err = listS3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    &testKey,
	})
	assert.Error(t, err)

	_, err = listS3Client.ListBucketInventoryConfigurations(ctx, &s3.ListBucketInventoryConfigurationsInput{
		Bucket: aws.String(bucketName),
	})
	assert.Error(t, err)
}

func TestAdminRole(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testKey := "testS3KeyName"
	body := "banana"
	terraformOptions := copyTerraformAndReturnOptions(t, "examples/simple", map[string]interface{}{})
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	bucketName := terraform.Output(t, terraformOptions, "bucket_name")

	versionId := CreateTestObject(t, ctx, bucketName, testKey)
	defer DeleteTestObject(t, ctx, bucketName, testKey, versionId)

	adminS3Client := S3ClientFromOutputArn(t, ctx, terraformOptions, "admin_role_arn")
	_, err := adminS3Client.ListBucketInventoryConfigurations(ctx, &s3.ListBucketInventoryConfigurationsInput{
		Bucket: aws.String(bucketName),
	})
	assert.NoError(t, err)

	_, err = adminS3Client.GetBucketAccelerateConfiguration(ctx, &s3.GetBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucketName),
	})
	assert.NoError(t, err)

	_, err = adminS3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(bucketName)})
	assert.NoError(t, err)

	_, err = adminS3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    &testKey,
	})
	assert.Error(t, err)

	_, err = adminS3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:     aws.String(bucketName),
		Key:        aws.String(testKey),
		Body:       strings.NewReader(body),
		ContentMD5: aws.String(md5EncodeBody(body)),
	})
	require.Error(t, err)

	_, err = adminS3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		VersionId: versionId,
		Bucket:    aws.String(bucketName),
		Key:       &testKey,
	})
	require.Error(t, err)
}

func TestMetadataRole(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	testKey := "testS3KeyName"
	body := "banana"
	terraformOptions := copyTerraformAndReturnOptions(t, "examples/simple", map[string]interface{}{})
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	bucketName := terraform.Output(t, terraformOptions, "bucket_name")

	versionId := CreateTestObject(t, ctx, bucketName, testKey)
	defer DeleteTestObject(t, ctx, bucketName, testKey, versionId)

	metadataS3Client := S3ClientFromOutputArn(t, ctx, terraformOptions, "metadata_role_arn")
	_, err := metadataS3Client.GetBucketAccelerateConfiguration(ctx, &s3.GetBucketAccelerateConfigurationInput{
		Bucket: aws.String(bucketName),
	})
	assert.NoError(t, err)

	_, err = metadataS3Client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(bucketName)})
	assert.Error(t, err)

	_, err = metadataS3Client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    &testKey,
	})
	assert.Error(t, err)

	_, err = metadataS3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:     aws.String(bucketName),
		Key:        aws.String(testKey),
		Body:       strings.NewReader(body),
		ContentMD5: aws.String(md5EncodeBody(body)),
	})
	require.Error(t, err)

	_, err = metadataS3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		VersionId: versionId,
		Bucket:    aws.String(bucketName),
		Key:       &testKey,
	})
	require.Error(t, err)
}

func TestUploadMultiPart(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	body := "large_multi_part_upload"
	testKey := "testS3KeyName"
	terraformOptions := copyTerraformAndReturnOptions(t, "examples/simple", map[string]interface{}{})
	defer terraform.Destroy(t, terraformOptions)

	terraform.InitAndApply(t, terraformOptions)

	bucketName := terraform.Output(t, terraformOptions, "bucket_name")

	s3Client := S3ClientFromOutputArn(t, ctx, terraformOptions, "write_role_arn")
	uploader := manager.NewUploader(s3Client, func(u *manager.Uploader) {
		u.PartSize = manager.MinUploadPartSize
	})
	_, err := uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:     &bucketName,
		Key:        &testKey,
		Body:       bytes.NewReader([]byte(body)),
		ContentMD5: aws.String(md5EncodeBody(body)),
	})
	require.NoError(t, err)
}

func TestCannotUseDifferentKeys(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	terraformOptions := copyTerraformAndReturnOptions(t, "examples/extra_key", map[string]interface{}{})
	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)

	bucketName := terraform.Output(t, terraformOptions, "bucket_name")
	bucketKeyArn := terraform.Output(t, terraformOptions, "bucket_kms_key_arn")
	bucketKeyId := terraform.Output(t, terraformOptions, "bucket_kms_key_id")
	additionalKey := terraform.Output(t, terraformOptions, "additional_kms_key_arn")

	var body string
	cfg := CreateConfig(t, ctx)
	client := s3.NewFromConfig(cfg)
	body = "different kms key"
	_, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               &bucketName,
		Key:                  aws.String("differentKey"),
		Body:                 strings.NewReader(body),
		ContentMD5:           aws.String(md5EncodeBody(body)),
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		SSEKMSKeyId:          &additionalKey})
	require.Error(t, err)
	body = "service key"
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               &bucketName,
		Key:                  aws.String("serviceKey"),
		Body:                 strings.NewReader(body),
		ContentMD5:           aws.String(md5EncodeBody(body)),
		ServerSideEncryption: types.ServerSideEncryptionAwsKms})
	require.Error(t, err)
	body = "AES"
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               &bucketName,
		Key:                  aws.String("aesKey"),
		Body:                 strings.NewReader(body),
		ContentMD5:           aws.String(md5EncodeBody(body)),
		ServerSideEncryption: types.ServerSideEncryptionAes256})
	require.Error(t, err)
	body = "Dont specify any"
	putOut, err := client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:     &bucketName,
		Key:        aws.String("noEncSpec"),
		Body:       strings.NewReader(body),
		ContentMD5: aws.String(md5EncodeBody(body))})
	require.NoError(t, err)
	assert.Equal(t, bucketKeyArn, *putOut.SSEKMSKeyId)
	body = "specify alg and bucket key"
	putOut, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               &bucketName,
		Key:                  aws.String("specKeyByArn"),
		Body:                 strings.NewReader(body),
		ContentMD5:           aws.String(md5EncodeBody(body)),
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		SSEKMSKeyId:          &bucketKeyArn})
	require.NoError(t, err)
	assert.Equal(t, bucketKeyArn, *putOut.SSEKMSKeyId)
	body = "specify alg and bucket key"
	putOut, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               &bucketName,
		Key:                  aws.String("specKeyById"),
		Body:                 strings.NewReader(body),
		ContentMD5:           aws.String(md5EncodeBody(body)),
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		SSEKMSKeyId:          &bucketKeyId})
	require.NoError(t, err)
	assert.Equal(t, bucketKeyArn, *putOut.SSEKMSKeyId)
}

func TestProvisionerCreateDestroy(t *testing.T) {
	t.Parallel()
	terraformOptions := copyTerraformAndReturnOptions(t, "examples/minimal", map[string]interface{}{})
	defer terraform.Destroy(t, terraformOptions)
	terraform.InitAndApply(t, terraformOptions)
}

func S3ClientFromOutputArn(t *testing.T, ctx context.Context, terraformOptions *terraform.Options, OutputArnName string) *s3.Client {
	cfg := CreateConfig(t, ctx)

	roleARN := terraform.Output(t, terraformOptions, OutputArnName)
	cfg = ConfigAssumeRole(t, cfg, roleARN)

	return s3.NewFromConfig(cfg)
}

func ConfigAssumeRole(t *testing.T, cfg aws.Config, roleARN string) aws.Config {
	provider := stscreds.NewAssumeRoleProvider(sts.NewFromConfig(cfg), roleARN)
	cfg.Credentials = aws.NewCredentialsCache(provider)
	_, Err := cfg.Credentials.Retrieve(context.Background())
	require.NoError(t, Err)
	return cfg
}

func CreateConfig(t *testing.T, ctx context.Context) aws.Config {
	cfg, Err := config.LoadDefaultConfig(
		ctx,
		// config.WithClientLogMode(aws.LogRetries | aws.LogRequest),
		config.WithRetryer(func() aws.Retryer {
			customRetry := retry.AddWithErrorCodes(retry.NewStandard(), "AccessDenied")
			customRetry = retry.AddWithMaxAttempts(customRetry, 7)
			return customRetry
		}),
		config.WithRegion(region),
	)
	require.NoError(t, Err)
	return cfg
}

func CreateTestObject(t *testing.T, ctx context.Context, bucketName string, testKey string) *string {
	terraformClient := s3.NewFromConfig(CreateConfig(t, ctx))
	body := "banana"
	uploadResponse, Err := terraformClient.PutObject(ctx, &s3.PutObjectInput{
		Bucket:     aws.String(bucketName),
		Key:        aws.String(testKey),
		Body:       strings.NewReader(body),
		ContentMD5: aws.String(md5EncodeBody(body))})
	require.NoError(t, Err)
	return uploadResponse.VersionId
}

func DeleteTestObject(t *testing.T, ctx context.Context, bucketName string, testKey string, versionId *string) {
	terraformClient := s3.NewFromConfig(CreateConfig(t, ctx))
	_, readDeleteErr := terraformClient.DeleteObject(ctx, &s3.DeleteObjectInput{
		VersionId: versionId,
		Bucket:    aws.String(bucketName),
		Key:       &testKey,
	})
	require.NoError(t, readDeleteErr)
}

func copyTerraformAndReturnOptions(t *testing.T, pathFromRootToSource string, additionalVars map[string]interface{}) *terraform.Options {
	testName := fmt.Sprintf("terratest-%s", strings.ToLower(random.UniqueId()))
	vars := map[string]interface{}{
		"test_name": testName,
	}
	for k, v := range additionalVars {
		vars[k] = v
	}
	return CopyTerraformAndReturnOptions(t, pathFromRootToSource, vars)
}

func CopyTerraformAndReturnOptions(t *testing.T, pathFromRootToSource string, vars map[string]interface{}) *terraform.Options {
	tempTestFolder := test_structure.CopyTerraformFolderToTemp(t, "..", pathFromRootToSource)
	log.Print(tempTestFolder)

	return terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: tempTestFolder,
		Vars:         vars,
	})
}

func md5EncodeBody(s string) string {
	hash := md5.Sum([]byte(s))
	var data []byte
	data = hash[:]
	return base64.StdEncoding.EncodeToString(data)
}
