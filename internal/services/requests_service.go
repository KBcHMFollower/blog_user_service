package services

import (
	"context"
	"database/sql"
	"errors"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	repositoriestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	servicestransfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/logger"
	dep "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces/dep"
)

type reqSvcRequestsStore interface {
	dep.RequestsCreator
	dep.RequestsGetter
}

type RequestsService struct {
	reqRepository reqSvcRequestsStore
	log           logger.Logger
}

func NewRequestsService(reqRepository reqSvcRequestsStore, log logger.Logger) *RequestsService {
	return &RequestsService{
		reqRepository: reqRepository,
		log:           log,
	}
}

func (rs *RequestsService) CheckAndCreate(ctx context.Context, checkInfo servicestransfer.RequestsCheckExistsInfo) (bool, error) {
	res, err := rs.reqRepository.Get(ctx, checkInfo.Key, nil)
	if err != nil && !errors.Is(err, ctxerrors.ErrNotFound) && !errors.Is(err, sql.ErrNoRows) {
		return false, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("cant get request from repository", err))
	}

	if res != nil {
		return true, nil
	}

	err = rs.reqRepository.Create(ctx, repositoriestransfer.CreateRequestInfo{
		Key: checkInfo.Key,
	}, nil)
	if err != nil {
		return false, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("cant create request in repository", err))
	}

	return false, nil
}
