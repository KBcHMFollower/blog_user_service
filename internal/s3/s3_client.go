package s3client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	ImageJpeg string = "image/jpeg"
)

type ImageStore interface {
	UploadFile(ctx context.Context, fileName string, fileBytes []byte, contentType string) (string, error)
}

type S3Client struct {
	minioClient *minio.Client
	bucketName  string
}

func New(endpoint string, accessKeyID string, secretAccessKey string, bucketName string) (*S3Client, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		return nil, err
	}

	return &S3Client{minioClient, bucketName}, nil
}

func (s *S3Client) UploadFile(ctx context.Context, fileName string, fileBytes []byte, contentType string) (string, error) {
	reader := bytes.NewReader(fileBytes)

	_, err := s.minioClient.PutObject(ctx, s.bucketName, fileName, reader, reader.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/%s/%s", s.minioClient.EndpointURL(), s.bucketName, fileName)
	return url, nil
}
