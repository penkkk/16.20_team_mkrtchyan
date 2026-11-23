package storage

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct{ client *s3.Client }

func NewS3(endpoint, region, accessKey, secretKey string) *S3 {
	return &S3{client: s3.New(s3.Options{Region: region, BaseEndpoint: aws.String(endpoint), UsePathStyle: true, Credentials: credentials.NewStaticCredentialsProvider(accessKey, secretKey, "")})}
}

func (s *S3) Put(ctx context.Context, bucket, key, contentType string, body io.ReadSeeker, size int64) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String(key), Body: body, ContentLength: aws.Int64(size), ContentType: aws.String(contentType)})
	return err
}
func (s *S3) Delete(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	return err
}

func (s *S3) Get(ctx context.Context, bucket, key string) (io.ReadCloser, string, int64, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(bucket), Key: aws.String(key)})
	if err != nil {
		return nil, "", 0, err
	}
	contentType := aws.ToString(result.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	return result.Body, contentType, aws.ToInt64(result.ContentLength), nil
}
