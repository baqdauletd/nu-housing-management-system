package database

import (
	"context"
	"log"
	"nu-housing-management-system/backend/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func ConnectMinIO(cfg *config.Config) (*minio.Client, error) {
	minioClient, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		log.Fatalln(err)
        return nil, err
	}

    ctx := context.Background()
    exists, err := minioClient.BucketExists(ctx, cfg.MinioBucket)
    if err != nil {
        log.Fatalln(err)
        return nil, err
    }
    if !exists {
        err = minioClient.MakeBucket(ctx, cfg.MinioBucket, minio.MakeBucketOptions{Region: "us-east-1"})
        if err != nil {
            log.Fatalln(err)
            return nil, err
        }
        log.Printf("Created MinIO bucket: %s\n", cfg.MinioBucket)
    } else {
        log.Printf("MinIO bucket %s already exists\n", cfg.MinioBucket)
    }

    return minioClient, nil
}
