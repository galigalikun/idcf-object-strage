package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

func main() {
	ctx := context.Background()
	var apiKey, secretKey, bucket, key string
	var timeout time.Duration

	flag.StringVar(&apiKey, "a", "", "api key.")
	flag.StringVar(&secretKey, "s", "", "secret key.")
	flag.StringVar(&bucket, "b", "", "Bucket name.")
	flag.StringVar(&key, "k", "", "Object key name.")
	flag.DurationVar(&timeout, "d", 0, "Upload timeout.")
	flag.Parse()

	cre := credentials.NewStaticCredentialsProvider(
		apiKey,
		secretKey,
		"")

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(cre),
		config.WithRegion("us-west-2"),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load configuration, %v\n", err)
		os.Exit(1)
	}

	svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
		o.BaseEndpoint = aws.String("ds.jp-east.idcfcloud.com")
		o.Region = "us-west-2"
	})

	var cancelFn func()
	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}
	defer cancelFn()

	if _, err := svc.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   os.Stdin,
	}); err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			code := apiErr.ErrorCode()
			message := apiErr.ErrorMessage()
			fmt.Fprintf(os.Stderr, "unexpected API error, code: %s, message: %s\n", code, message)
		}
		os.Exit(1)
	}

	fmt.Printf("successfully uploaded file to %s/%s\n", bucket, key)
}
