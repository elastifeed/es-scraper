package storage

import (
	"bytes"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	uuid "github.com/satori/go.uuid"
)

type s3driver struct {
	client   *s3.S3
	subdir   string
	endpoint string
}

// NewS3 creates a new S3 Storage driver based on a given config
func NewS3(awsconf *aws.Config, subdir string, endpoint string) (Storager, error) {
	var c s3driver
	c.endpoint = endpoint
	c.subdir = subdir
	c.client = s3.New(session.New(awsconf))

	// Create storage bucket, public readable. Maybe change that @TODO
	_, err := c.client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(subdir),
	})

	if err != nil {
		return nil, err
	}

	log.Printf("Created S3 bucket %s", subdir)

	return c, nil
}

// Upload a byte slice to the S3 provider and output a public accessable URL for it
func (i s3driver) Upload(data []byte, ending string) (string, error) {
	key := uuid.NewV4().String() + "." + ending

	_, err := i.client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(i.subdir),
		Key:    aws.String(key),
		ACL:    aws.String("public-read"),
		Body:   bytes.NewReader(data),
	})

	if err != nil {
		return "", err
	}

	url := i.endpoint + i.subdir + "/" + key

	log.Printf("Uploaded %d bytes to S3: %s", len(data), url)

	return url, nil
}
