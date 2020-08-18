package bzserver

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
func Serve(log *log.Logger, bucket, key string) (*s3.GetObjectOutput, error) {
	s3Config := &aws.Config{
		Credentials: credentials.NewStaticCredentials(
			config.File().GetString("backblaze.keyID"),
			config.File().GetString("backblaze.applicationKey"),
			config.File().GetString("backblaze.token"),
		),
		Endpoint:         aws.String(config.File().GetString("backblaze.s3.endpoint")),
		Region:           aws.String("eu-central-003"),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession := session.New(s3Config)
	s3Client := s3.New(newSession)
	res, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %v", err)
	}
	log.Printf("downloaded file %s\n", key)

	return res, nil
}
