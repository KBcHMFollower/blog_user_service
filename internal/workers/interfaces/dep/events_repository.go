package workers_dep

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/KBcHMFollower/blog_user_service/internal/repository"
	"github.com/google/uuid"
)

type EventUpdater interface {
	SetStatusInEvents(ctx context.Context, eventsId []uuid.UUID, status repository.MessageStatus) error
}

type EventGetter interface {
	GetEventsWithStatus(ctx context.Context, status repository.MessageStatus, limit uint64) ([]*models.EventInfo, error)
}
