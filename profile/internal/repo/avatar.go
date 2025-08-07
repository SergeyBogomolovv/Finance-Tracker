package repo

import (
	"context"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	avatarsBucket = "avatars"
)

type avatarRepo struct {
	uploader *manager.Uploader
	client   *s3.Client
}

func NewAvatarRepo(conf aws.Config) *avatarRepo {
	client := s3.NewFromConfig(conf, func(o *s3.Options) {
		o.UsePathStyle = true
	})
	manager := manager.NewUploader(client)

	return &avatarRepo{
		client:   client,
		uploader: manager,
	}
}

func (r *avatarRepo) Upload(ctx context.Context, key string, data io.Reader) error {
	_, err := r.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String(avatarsBucket),
		Key:    aws.String(key),
		Body:   data,
	})
	return err
}
