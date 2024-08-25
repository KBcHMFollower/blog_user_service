package services_dep_interfaces

import (
	"context"
	"database/sql"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type RequestsCreator interface {
	Create(ctx context.Context, info repositories_transfer.CreateRequestInfo, tx *sql.Tx) error
}

type RequestsGetter interface {
	Get(ctx context.Context, key uuid.UUID, tx *sql.Tx) (*models.Request, error)
}
