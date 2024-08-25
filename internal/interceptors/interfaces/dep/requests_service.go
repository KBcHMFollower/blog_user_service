package dep

import (
	"context"
	services_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
)

type IdempotencyKeysChecker interface {
	CheckAndCreate(ctx context.Context, checkInfo services_transfer.RequestsCheckExistsInfo) (bool, error)
}
