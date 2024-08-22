package s3client

import "context"

type S3Client interface {
	UploadFile(ctx context.Context, fileName string, fileBytes []byte, contentType string) (string, error)
	GetFile(ctx context.Context, fileName string) ([]byte, error)
	Stop() error
}
