package database

import (
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    "nu-housing-management-system/backend/internal/config"
)

func ConnectMinIO(cfg *config.Config) (*minio.Client, error) {
    return minio.New(cfg.MinioEndpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
        Secure: false, // true if using HTTPS
    })
}
