package database

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"nu-housing-management-system/backend/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStore struct {
	Client        *minio.Client
	PresignClient *minio.Client
	Bucket        string
}

func ConnectMinIO(cfg *config.Config) (*MinIOStore, error) {
	minioClient, err := newMinIOClient(cfg.MinioEndpoint, cfg.MinioUseSSL, cfg.MinioAccessKey, cfg.MinioSecretKey)
	if err != nil {
		return nil, fmt.Errorf("create MinIO client: %w", err)
	}

	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, cfg.MinioBucket)
	if err != nil {
		return nil, fmt.Errorf("check MinIO bucket %q: %w", cfg.MinioBucket, err)
	}
	if !exists {
		err = minioClient.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{Region: "us-east-1"})
		if err != nil {
			return nil, fmt.Errorf("create MinIO bucket %q: %w", cfg.MinioBucket, err)
		}
	}

	presignEndpoint := cfg.MinioPublicEndpoint
	if presignEndpoint == "" {
		presignEndpoint = cfg.MinioEndpoint
	}

	presignClient, err := newMinIOClient(presignEndpoint, cfg.MinioUseSSL, cfg.MinioAccessKey, cfg.MinioSecretKey)
	if err != nil {
		return nil, fmt.Errorf("create MinIO presign client: %w", err)
	}

	return &MinIOStore{
		Client:        minioClient,
		PresignClient: presignClient,
		Bucket:        cfg.MinioBucket,
	}, nil
}

func newMinIOClient(endpoint string, useSSL bool, accessKey string, secretKey string) (*minio.Client, error) {
	normalizedEndpoint, secure, err := normalizeMinIOEndpoint(endpoint, useSSL)
	if err != nil {
		return nil, err
	}

	return minio.New(normalizedEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: secure,
	})
}

func normalizeMinIOEndpoint(raw string, useSSL bool) (string, bool, error) {
	endpoint := strings.TrimSpace(raw)
	if endpoint == "" {
		return "", useSSL, fmt.Errorf("empty MinIO endpoint")
	}

	if strings.Contains(endpoint, "://") {
		parsed, err := url.Parse(endpoint)
		if err != nil {
			return "", useSSL, fmt.Errorf("parse MinIO endpoint %q: %w", raw, err)
		}
		if parsed.Host == "" {
			return "", useSSL, fmt.Errorf("invalid MinIO endpoint %q", raw)
		}

		return parsed.Host, parsed.Scheme == "https", nil
	}

	return endpoint, useSSL, nil
}
