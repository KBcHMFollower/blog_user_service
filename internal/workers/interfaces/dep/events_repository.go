package workers_dep

import (
	"context"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
)

type EventUpdater interface {
	Update(ctx context.Context, info repositoriestransfer.UpdateEventInfo) error
	UpdateMany(ctx context.Context, info repositoriestransfer.UpdateManyEventInfo) error
}

type EventGetter interface {
	Events(ctx context.Context, info repositoriestransfer.GetEventsInfo) ([]*models.EventInfo, error)
}
