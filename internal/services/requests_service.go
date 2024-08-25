package services

import (
	"context"
	"github.com/KBcHMFollower/blog_user_service/internal/domain"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	services_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/services"
	"github.com/KBcHMFollower/blog_user_service/internal/repository"
	dep "github.com/KBcHMFollower/blog_user_service/internal/services/interfaces/dep"
	"log/slog"
)

type RequestsStore interface {
	dep.RequestsCreator
	dep.RequestsGetter
}

type RequestsService struct {
	reqRepository repository.RequestsStore
	log           *slog.Logger
}

func NewRequestsService(reqRepository repository.RequestsStore, log *slog.Logger) *RequestsService {
	return &RequestsService{
		reqRepository: reqRepository,
		log:           log,
	}
}

// TODO: ДОБАВИТЬ ИНТЕРЦЕПТОРЫ С ЭТИМ МЕТОДОМ
func (rs *RequestsService) CheckAndCreate(ctx context.Context, checkInfo services_transfer.RequestsCheckExistsInfo) (bool, error) {
	op := "RequestsService.CheckAndCreate"

	log := rs.log.With(
		slog.String("op", op),
	)

	res, err := rs.reqRepository.Get(ctx, checkInfo.Key, nil)
	if err != nil {
		log.Error(err.Error())
		return false, domain.AddOpInErr(err, op)
	}

	if res != nil {
		return true, nil
	}

	err = rs.reqRepository.Create(ctx, repositories_transfer.CreateRequestInfo{
		Key: checkInfo.Key,
	}, nil)
	if err != nil {
		log.Error(err.Error())
		return false, domain.AddOpInErr(err, op)
	}

	return false, nil
}
