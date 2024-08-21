package services_dep_interfaces

import "context"

type ImageGetter interface {
	GetFile(ctx context.Context, fileName string) ([]byte, error)
}

type ImageUploader interface {
	UploadFile(ctx context.Context, fileName string, fileBytes []byte, contentType string) (string, error)
}
