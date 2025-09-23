// pkg/minio/minio.go
package minio

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"document-embeddings/internal/config"
)

type Client struct {
	*minio.Client
	BucketName string
}

func New(cfg config.MinIOConfig) (*Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Client{
		Client:     client,
		BucketName: cfg.BucketName,
	}, nil
}

func (c *Client) GetObject(ctx context.Context, objectPath string) (io.ReadCloser, error) {
	return c.Client.GetObject(ctx, c.BucketName, objectPath, minio.GetObjectOptions{})
}

func (c *Client) PutObject(ctx context.Context, objectPath string, reader io.Reader, objectSize int64, opts minio.PutObjectOptions) (minio.UploadInfo, error) {
	return c.Client.PutObject(ctx, c.BucketName, objectPath, reader, objectSize, opts)
}
