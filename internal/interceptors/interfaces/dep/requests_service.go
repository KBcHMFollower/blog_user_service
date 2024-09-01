package dep

import (
	"context"
	servicestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
)

type IdempotencyKeysChecker interface {
	CheckAndCreate(ctx context.Context, checkInfo servicestransfer.RequestsCheckExistsInfo) (bool, error)
}
