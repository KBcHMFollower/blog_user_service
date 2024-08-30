package workers_dep

import (
	"context"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	"github.com/google/uuid"
)

type EventUpdater interface {
	SetStatuses(ctx context.Context, eventsId []uuid.UUID, status repositories_transfer.MessageStatus) error
}

type EventGetter interface {
	EventsWithStatus(ctx context.Context, status repositories_transfer.MessageStatus, limit uint64) ([]*models.EventInfo, error)
}
