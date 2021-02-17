package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
)

// Retrieve a file from the backblaze cloude storrage.
func retrieve(app *app, objectKey string) (*s3.GetObjectOutput, error) {
	obj := s3.GetObjectInput{
		Bucket: &app.bucket,
		Key:    &objectKey,
	}
	cfg := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			app.keyID,
			app.appKey,
			app.token,
		),
		Endpoint:         aws.String(app.s3),
		Region:           aws.String(app.region),
		S3ForcePathStyle: aws.Bool(true),
	}
	awsSession := session.New(cfg)
	res, err := s3.New(awsSession).GetObject(&obj)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("retrieving object key: %v", objectKey))
	}

	return res, nil
}
