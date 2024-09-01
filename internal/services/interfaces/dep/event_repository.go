package services_dep_interfaces

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type EventGetter interface {
	Event(ctx context.Context, eventId uuid.UUID) (*models.EventInfo, error)
}

type EventCreator interface {
	Create(ctx context.Context, info repositoriestransfer.CreateEventInfo, tx database.Transaction) error
}
