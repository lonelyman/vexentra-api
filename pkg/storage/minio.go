package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"vexentra-api/internal/config"
	"vexentra-api/pkg/logger"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type minioStorage struct {
	client *minio.Client
	bucket string
	region string
	log    logger.Logger
}

func NewMinioStorage(cfg config.StorageConfig, l logger.Logger) (ObjectStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}
	return &minioStorage{
		client: client,
		bucket: cfg.Bucket,
		region: cfg.Region,
		log:    l,
	}, nil
}

func (s *minioStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	return s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{
		Region:        s.region,
		ObjectLocking: false,
	})
}

func (s *minioStorage) PresignPut(ctx context.Context, objectKey, contentType string, expires time.Duration) (*PresignedUpload, error) {
	u, err := s.client.PresignedPutObject(ctx, s.bucket, objectKey, expires)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{}
	if contentType != "" {
		headers["Content-Type"] = contentType
	}
	return &PresignedUpload{
		URL:     u.String(),
		Headers: headers,
	}, nil
}

func (s *minioStorage) PresignGet(ctx context.Context, objectKey string, expires time.Duration) (string, error) {
	u, err := s.client.PresignedGetObject(ctx, s.bucket, objectKey, expires, nil)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

func (s *minioStorage) StatObject(ctx context.Context, objectKey string) (*ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, s.bucket, objectKey, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{
		Size:        info.Size,
		ETag:        info.ETag,
		ContentType: info.ContentType,
	}, nil
}

func (s *minioStorage) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (s *minioStorage) CopyObject(ctx context.Context, srcObjectKey, dstObjectKey, contentType string) error {
	src := minio.CopySrcOptions{
		Bucket: s.bucket,
		Object: srcObjectKey,
	}
	dst := minio.CopyDestOptions{
		Bucket: s.bucket,
		Object: dstObjectKey,
	}
	if contentType != "" {
		dst.UserMetadata = map[string]string{
			"Content-Type": contentType,
		}
	}
	_, err := s.client.CopyObject(ctx, dst, src)
	return err
}

func (s *minioStorage) RemoveObject(ctx context.Context, objectKey string) error {
	err := s.client.RemoveObject(ctx, s.bucket, objectKey, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("remove object %s failed: %w", objectKey, err)
	}
	return nil
}
