package workers_dep

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type EventUpdater interface {
	SetSentStatusesInEvents(ctx context.Context, eventsId []uuid.UUID) error
}

type EventGetter interface {
	GetEvents(ctx context.Context, filterTarget string, filterValue interface{}, limit uint64) ([]*models.EventInfo, error)
}
