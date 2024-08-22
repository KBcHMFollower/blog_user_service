package s3client

import (
	"bytes"
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"io"
	"log"
)

const (
	ImageJpeg string = "image/jpeg"
)

type MinioClient struct {
	minioClient *minio.Client
	bucketName  string
}

func NewMinioClient(endpoint string, accessKeyID string, secretAccessKey string, bucketName string) (*MinioClient, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Println("Error initializing MinIO client:", err)
		return nil, err
	}

	err = ensureBucketExists(minioClient, bucketName)
	if err != nil {
		return nil, err
	}

	err = setBucketPolicy(minioClient, bucketName)
	if err != nil {
		return nil, err
	}

	return &MinioClient{minioClient, bucketName}, nil
}

func (s *MinioClient) UploadFile(ctx context.Context, fileName string, fileBytes []byte, contentType string) (string, error) {
	reader := bytes.NewReader(fileBytes)

	_, err := s.minioClient.PutObject(ctx, s.bucketName, fileName, reader, reader.Size(), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		fmt.Println("Error uploading file: ", err)
		return "", err
	}

	fileURL := fmt.Sprintf("http://%s/%s/%s", s.minioClient.EndpointURL().Host, s.bucketName, fileName)

	return fileURL, nil
}

func (s *MinioClient) GetFile(ctx context.Context, fileName string) ([]byte, error) {
	object, err := s.minioClient.GetObject(ctx, s.bucketName, fileName, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, object); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (s *MinioClient) Stop() error {
	//TODO: НЕ НАШЕЛ МЕТОД STOP

	return nil
}

func ensureBucketExists(client *minio.Client, bucketName string) error {
	exists, err := client.BucketExists(context.Background(), bucketName)
	if err != nil {
		return err
	}
	if !exists {
		err = client.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	} else {
	}
	return nil
}

func setBucketPolicy(client *minio.Client, bucketName string) error {
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {
					"AWS": "*"
				},
				"Action": [
					"s3:GetObject"
				],
				"Resource": [
					"arn:aws:s3:::%s/*"
				]
			}
		]
	}`, bucketName)

	err := client.SetBucketPolicy(context.Background(), bucketName, policy)
	if err != nil {
		return err
	}
	return nil
}
