package main

import (
	"context"
	"testing"
)

func TestNewS3Client(t *testing.T) {
	// Use dummy credentials and disable machine-specific profiles and metadata.
	t.Setenv("AWS_ACCESS_KEY_ID", "environment-access")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "environment-secret")
	t.Setenv("AWS_SESSION_TOKEN", "environment-token")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_CONFIG_FILE", t.TempDir()+"/config")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", t.TempDir()+"/credentials")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")

	for _, tt := range []struct {
		name, access, secret, wantAccess, wantSecret, wantToken string
	}{
		{"environment", "", "", "environment-access", "environment-secret", "environment-token"},
		{"flags override environment", "flag-access", "flag-secret", "flag-access", "flag-secret", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client, err := newS3Client(ctx, tt.access, tt.secret)
			if err != nil {
				t.Fatal(err)
			}
			opts := client.Options()
			if opts.BaseEndpoint == nil || *opts.BaseEndpoint != "https://ds.jp-east.idcfcloud.com" {
				t.Fatal("expected the IDCF HTTPS endpoint")
			}
			if !opts.UsePathStyle {
				t.Fatal("expected path-style S3 requests")
			}
			creds, err := opts.Credentials.Retrieve(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if creds.AccessKeyID != tt.wantAccess || creds.SecretAccessKey != tt.wantSecret || creds.SessionToken != tt.wantToken {
				t.Fatal("unexpected credentials source")
			}
		})
	}
}

func TestNewS3ClientRejectsIncompleteCredentials(t *testing.T) {
	for _, pair := range [][2]string{{"access", ""}, {"", "secret"}} {
		client, err := newS3Client(context.Background(), pair[0], pair[1])
		if err == nil || client != nil {
			t.Fatal("expected incomplete credentials to be rejected before loading config")
		}
	}
}
