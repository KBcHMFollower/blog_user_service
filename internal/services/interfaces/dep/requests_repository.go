package services_dep_interfaces

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type RequestsCreator interface {
	Create(ctx context.Context, info repositoriestransfer.CreateRequestInfo, tx database.Transaction) error
}

type RequestsGetter interface {
	Get(ctx context.Context, key uuid.UUID, tx database.Transaction) (*models.Request, error)
}
