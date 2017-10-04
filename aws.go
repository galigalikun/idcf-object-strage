package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/private/protocol/restxml"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/benmcclelland/s3v2"
)

func main() {
	var apiKey, secretKey, bucket, key string
	var timeout time.Duration

	flag.StringVar(&apiKey, "a", "", "api key.")
	flag.StringVar(&secretKey, "s", "", "secret key.")
	flag.StringVar(&bucket, "b", "", "Bucket name.")
	flag.StringVar(&key, "k", "", "Object key name.")
	flag.DurationVar(&timeout, "d", 0, "Upload timeout.")
	flag.Parse()

	cre := credentials.NewStaticCredentials(
		apiKey,
		secretKey,
		"")

	sess := session.Must(session.NewSession())
	svc := s3.New(sess, &aws.Config{
		Credentials:      cre,
		Endpoint:         aws.String("ds.jp-east.idcfcloud.com"),
		Region:           aws.String(endpoints.UsWest2RegionID),
		S3ForcePathStyle: aws.Bool(true),
		LogLevel:         aws.LogLevel(aws.LogDebugWithRequestErrors | aws.LogDebug | aws.LogDebugWithSigning),
	})
	svc.Handlers.Sign.Clear()
	// svc.Handlers.Sign.PushBackNamed(corehandlers.BuildContentLengthHandler)
	// svc.Handlers.Sign.PushBackNamed(v2.SignRequestHandler)
	svc.Handlers.Sign.PushBackNamed(s3v2.SignRequestHandler)
	svc.Handlers.Build.PushBackNamed(restxml.BuildHandler)
	svc.Handlers.Unmarshal.PushBackNamed(restxml.UnmarshalHandler)
	svc.Handlers.UnmarshalMeta.PushBackNamed(restxml.UnmarshalMetaHandler)
	svc.Handlers.UnmarshalError.PushBackNamed(restxml.UnmarshalErrorHandler)

	ctx := context.Background()
	var cancelFn func()
	if timeout > 0 {
		ctx, cancelFn = context.WithTimeout(ctx, timeout)
	}
	defer cancelFn()

	_, err := svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   os.Stdin,
	})
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok && aerr.Code() == request.CanceledErrorCode {
			// If the SDK can determine the request or retry delay was canceled
			// by a context the CanceledErrorCode error code will be returned.
			fmt.Fprintf(os.Stderr, "upload canceled due to timeout, %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "failed to upload object, %v\n", err)
		}
		os.Exit(1)
	}

	fmt.Printf("successfully uploaded file to %s/%s\n", bucket, key)
}
