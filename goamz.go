package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mitchellh/goamz/aws"
	"github.com/mitchellh/goamz/s3"
)

func main() {
	var apiKey, secretKey string
	flag.StringVar(&apiKey, "a", "", "api key.")
	flag.StringVar(&secretKey, "s", "", "secret key.")
	flag.Parse()

	auth, err := aws.GetAuth(apiKey, secretKey)
	if err != nil {
		fmt.Printf("err: %v\n", err)
		os.Exit(-1)
	}

	// Open Bucket
	s := s3.New(auth, aws.Region{
		Name:                 "idcf",
		S3Endpoint:           "https://ds.jp-east.idcfcloud.com",
		S3BucketEndpoint:     "",
		S3LocationConstraint: true,
		S3LowercaseBucket:    true,
	})

	bucket := s.Bucket("alice")

	data := []byte("hoge")
	err = bucket.Put("sample.txt", data, "text/plain", s3.PublicRead)
	if err != nil {
		fmt.Printf("put err:%v", err)
	}
}
