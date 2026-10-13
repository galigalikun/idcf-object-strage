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
		fmt.Fprintf(os.Stderr, "Usage: %s -b BUCKET -k KEY [-d DURATION] < FILE\nCredentials: AWS_ACCESS_KEY_ID / AWS_SECRET_ACCESS_KEY or an AWS profile; -a / -s override them.\n", os.Args[0])
		flag.PrintDefaults()
	}
}

func newS3Client(ctx context.Context, accessKey, secret string) (*s3.Client, error) {
	if (accessKey == "") != (secret == "") {
		return nil, errors.New("both -a and -s must be provided together")
	}

	opts := []func(*config.LoadOptions) error{
		config.WithBaseEndpoint("https://ds.jp-east.idcfcloud.com"),
		config.WithRegion("us-west-2"),
	}
	if accessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secret, "")))
	}

	cfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	}), nil
}

func run() error {
	ctx := context.Background()

	svc, err := newS3Client(ctx, *apiKey, *secretKey)
	if err != nil {
		return err
	}

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
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("successfully uploaded file to %s/%s\n", *bucket, *key)
}
