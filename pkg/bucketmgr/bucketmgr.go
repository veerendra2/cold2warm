package bucketmgr

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type Config struct {
	Region     string `name:"region" help:"The region where the S3 bucket is hosted (e.g., nl-ams)." env:"REGION" default:"nl-ams"`
	Endpoint   string `name:"endpoint" help:"Custom S3 endpoint URL (e.g., s3.nl-ams.scw.cloud). Do NOT include the bucket name." env:"ENDPOINT" required:""`
	AccessKey  string `name:"access-key" help:"The access key ID for S3 authentication." env:"ACCESS_KEY" required:""`
	SecretKey  string `name:"secret-key" help:"The secret access key for S3 authentication." env:"SECRET_KEY" required:""`
	BucketName string `name:"bucket-name" help:"The name of the target S3 bucket." env:"BUCKET_NAME" required:""`

	Days         int32  `name:"days" help:"Number of days to keep the restored object" env:"RESTORE_DAYS" default:"30"`
	ObjectPrefix string `name:"prefix" help:"Filter objects by this prefix (e.g., 'backups/')." env:"OBJECT_PREFIX" default:""`
}

type Client interface {
	RestoreObject(ctx context.Context, object string) error
	ListObjectsPaginator(ctx context.Context) (*s3.ListObjectsV2Paginator, error)
}

type client struct {
	s3Client s3.Client

	bucketName   string
	days         int32
	objectPrefix string
}

func (c *client) RestoreObject(ctx context.Context, object string) error {
	_, err := c.s3Client.RestoreObject(ctx, &s3.RestoreObjectInput{
		Bucket: aws.String(c.bucketName),
		Key:    aws.String(object),
		RestoreRequest: &types.RestoreRequest{
			Days: aws.Int32(c.days),
		}},
	)
	if err != nil {
		return err
	}

	return nil
}

func (c *client) ListObjectsPaginator(ctx context.Context) (*s3.ListObjectsV2Paginator, error) {
	maxKey := int32(32)

	paginator := s3.NewListObjectsV2Paginator(&c.s3Client, &s3.ListObjectsV2Input{
		Bucket:  &c.bucketName,
		Prefix:  &c.objectPrefix,
		MaxKeys: &maxKey,
	})

	return paginator, nil
}

func NewClient(ctx context.Context, cfg Config) (Client, error) {
	s3Config, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKey,
				cfg.SecretKey, "",
			),
		),
	)
	if err != nil {
		return nil, err
	}

	endpoint := cfg.Endpoint
	if !strings.HasPrefix(endpoint, "http") {
		endpoint = fmt.Sprintf("https://%s", endpoint)
	}

	s3Client := s3.NewFromConfig(s3Config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
	})

	return &client{
		s3Client: *s3Client,

		bucketName:   cfg.BucketName,
		days:         cfg.Days,
		objectPrefix: cfg.ObjectPrefix,
	}, nil
}
