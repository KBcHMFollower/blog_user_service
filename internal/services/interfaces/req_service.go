package services_interfaces

import (
	"context"
	servicestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
)

type ReqService interface {
	CheckAndCreate(ctx context.Context, checkInfo servicestransfer.RequestsCheckExistsInfo) (bool, error)
}
