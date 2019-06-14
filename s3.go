package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

const (
	endpoint = "http://localhost:30098"
)

func newBool() *bool {
	b := true
	return &b
}

func main() {
	// Credentials: &credentials.NewStaticCredentials("K279UGQBCW1RM3G1IITH", "s11DBCiqv9hnqoJ9drpEAQJkkBO2EP0Gv7u6MgLf", ""),
	c := s3.New(session.New(&aws.Config{
		Credentials:      credentials.NewStaticCredentials("K279UGQBCW1RM3G1IITH", "s11DBCiqv9hnqoJ9drpEAQJkkBO2EP0Gv7u6MgLf", ""),
		Endpoint:         aws.String("http://localhost:30098"),
		Region:           aws.String("us-east-1"), // Somehow this is needed
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
	}))

	_, err := c.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String("hellofromgolang"),
	})

	if err != nil {
		log.Fatal(err)
	}

	_, err = c.PutObject(&s3.PutObjectInput{
		Bucket: aws.String("hellofromgolang"),
		Key:    aws.String("testkey"),
		Body:   strings.NewReader("Is this working? :-)"),
		ACL:    aws.String("public-read"),
	})

	if err != nil {
		log.Fatal(err)
	}

	data, err := c.GetObject(&s3.GetObjectInput{
		Bucket: aws.String("hellofromgolang"),
		Key:    aws.String("testkey"),
	})

	if err != nil {
		log.Fatal(err)
	}

	b, _ := ioutil.ReadAll(data.Body)
	fmt.Println(string(b))

}
