package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
)

var (
	apiKey    = flag.String("a", "", "api key.")
	secretKey = flag.String("s", "", "secret key.")
	timeout   = flag.Duration("d", 0, "Upload timeout.")
	bucket    = flag.String("b", "", "Bucket name.")
	key       = flag.String("k", "", "Object key name.")
)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s -a API_KEY -s SECRET_KEY -b BUCKET -k KEY [-d DURATION] < FILE\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()
}

func run() error {
	ctx := context.Background()

	cre := credentials.NewStaticCredentialsProvider(
		*apiKey,
		*secretKey,
		"")

	cfg, err := config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(cre),
		config.WithBaseEndpoint("ds.jp-east.idcfcloud.com"),
		config.WithRegion("us-west-2"),
	)
	if err != nil {
		return err
	}

	svc := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	var cancelFn func()
	if *timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, *timeout)
		defer cancelFn()
	}

	if _, err := svc.PutObject(ctx, &s3.PutObjectInput{
		Bucket: bucket,
		Key:    key,
		Body:   os.Stdin,
	}); err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			code := apiErr.ErrorCode()
			message := apiErr.ErrorMessage()
			return fmt.Errorf("unexpected API error, code: %s, message: %s", code, message)
		}
		return err
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("successfully uploaded file to %s/%s\n", *bucket, *key)
}
