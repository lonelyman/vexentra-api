package storage

import (
	"context"
	"io"
	"time"
)

type PresignedUpload struct {
	URL     string
	Headers map[string]string
}

type ObjectInfo struct {
	Size        int64
	ETag        string
	ContentType string
}

type ObjectStorage interface {
	EnsureBucket(ctx context.Context) error
	PresignPut(ctx context.Context, objectKey, contentType string, expires time.Duration) (*PresignedUpload, error)
	PresignGet(ctx context.Context, objectKey string, expires time.Duration) (string, error)
	StatObject(ctx context.Context, objectKey string) (*ObjectInfo, error)
	GetObject(ctx context.Context, objectKey string) (io.ReadCloser, error)
	CopyObject(ctx context.Context, srcObjectKey, dstObjectKey, contentType string) error
	RemoveObject(ctx context.Context, objectKey string) error
}
