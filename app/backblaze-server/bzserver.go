package main

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/packago/config"
)

// Serve a file from the backblaze cloude storrage.
func serve(log *log.Logger, bucket, key string) (*s3.GetObjectOutput, error) {
	obj := s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}
	cfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			config.File().GetString("backblaze.keyID"),
			config.File().GetString("backblaze.applicationKey"),
			config.File().GetString("backblaze.token"),
		),
		Endpoint:         aws.String(config.File().GetString("backblaze.s3.endpoint")),
		Region:           aws.String("eu-central-003"),
		S3ForcePathStyle: aws.Bool(true),
	}
	awsSession := session.New(cfg)
	res, err := s3.New(awsSession).GetObject(&obj)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}
	log.Printf("downloaded file %s\n", key)

	return res, nil
}
