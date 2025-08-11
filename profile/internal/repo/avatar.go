package repo

import (
	"FinanceTracker/profile/internal/config"
	"FinanceTracker/profile/pkg/logger"
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"

	_ "golang.org/x/image/webp"
	_ "image/gif"
	_ "image/png"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type avatarRepo struct {
	uploader *manager.Uploader
	client   *s3.Client
	bucket   string
}

func MustAvatarRepo(ctx context.Context, conf config.S3) *avatarRepo {
	cfg, err := awsConf.LoadDefaultConfig(ctx,
		awsConf.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(conf.AccessKey, conf.SecretKey, "")),
		awsConf.WithBaseEndpoint(conf.Endpoint),
		awsConf.WithRegion(conf.Region),
	)
	if err != nil {
		logger.Error(ctx, "failed to load AWS config", "err", err)
		os.Exit(1)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	manager := manager.NewUploader(client)

	return &avatarRepo{
		client:   client,
		uploader: manager,
		bucket:   conf.Bucket,
	}
}

func (r *avatarRepo) Upload(ctx context.Context, userID int, data io.Reader) (string, error) {
	img, _, err := image.Decode(data)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return "", fmt.Errorf("failed to encode image: %w", err)
	}

	avatarID := fmt.Sprintf("avatars/%d.jpg", userID)

	_, err = r.uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(avatarID),
		Body:        bytes.NewReader(buf.Bytes()),
		ContentType: aws.String("image/jpeg"),
	})
	if err != nil {
		return "", err
	}
	return avatarID, nil
}
