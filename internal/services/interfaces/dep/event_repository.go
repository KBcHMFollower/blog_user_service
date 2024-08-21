package services_dep_interfaces

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type EventGetter interface {
	GetEventById(ctx context.Context, eventId uuid.UUID) (*models.EventInfo, error)
	GetEvents(ctx context.Context, filterTarget string, filterValue interface{}, limit uint64) ([]*models.EventInfo, error)
}
