package services

import (
	"context"
	ctxerrors "github.com/KBcHMFollower/blog_user_service/internal/domain/errors"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	services_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	dep "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces/dep"
	"log/slog"
)

type RequestsStore interface {
	dep.RequestsCreator
	dep.RequestsGetter
}

type RequestsService struct {
	reqRepository RequestsStore
	log           *slog.Logger
}

func NewRequestsService(reqRepository RequestsStore, log *slog.Logger) *RequestsService {
	return &RequestsService{
		reqRepository: reqRepository,
		log:           log,
	}
}

func (rs *RequestsService) CheckAndCreate(ctx context.Context, checkInfo services_transfer.RequestsCheckExistsInfo) (bool, error) {
	res, err := rs.reqRepository.Get(ctx, checkInfo.Key, nil)
	if err != nil {
		return false, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("cant get request from repository", err))
	}

	if res != nil {
		return true, nil
	}

	err = rs.reqRepository.Create(ctx, repositories_transfer.CreateRequestInfo{
		Key: checkInfo.Key,
	}, nil)
	if err != nil {
		return false, ctxerrors.WrapCtx(ctx, ctxerrors.Wrap("cant create request in repository", err))
	}

	return false, nil
}
