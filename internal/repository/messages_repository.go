package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/KBcHMFollower/blog_user_service/internal/database"
	repositories_transfer "github.com/KBcHMFollower/blog_user_service/internal/domain/layers_TOs/repositories"
	"github.com/KBcHMFollower/blog_user_service/internal/domain/models"
	rep_utils "github.com/KBcHMFollower/blog_user_service/internal/repository/lib"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

const (
	messagesTable = "amqp_messages"
)

type MessageStatus string

const (
	MessagesStatusCol    = "status"
	messagesEventIdCol   = "event_id"
	messagesAllCol       = "*"
	messagesEventTypeCol = "event_type"
	messagesPayloadCol   = "payload"
)

const (
	MessagesSentStatus    MessageStatus = "sent"
	MessagesErrorStatus   MessageStatus = "error"
	MessagesWaitingStatus MessageStatus = "waiting"
)

type EventFilter struct {
}

type EventRepository struct {
	db database.DBWrapper
}

func NewEventRepository(dbDriver database.DBWrapper) *EventRepository {
	return &EventRepository{db: dbDriver}
}

func (r *EventRepository) GetEventsWithStatus(ctx context.Context, status MessageStatus, limit uint64) ([]*models.EventInfo, error) {
	op := "UserRepository.getSubInfo"

	builder := rep_utils.QBuilder.PHFormat(squirrel.Dollar)
	eventInfos := make([]*models.EventInfo, 0)

	query := builder.
		Sel(messagesAllCol).
		Frm(messagesTable).
		Wr(squirrel.Eq{MessagesStatusCol: status}).
		Lim(limit)

	if err := query.QryCtx(ctx, r.db, &eventInfos); err != nil {
		return eventInfos, fmt.Errorf("%s : failed to execute query: %w", op, err)
	}

	return eventInfos, nil
}

func (r *EventRepository) SetStatusInEvents(ctx context.Context, eventsId []uuid.UUID, status MessageStatus) error {
	op := "UserRepository.getSubInfo"

	builder := rep_utils.QBuilder.PHFormat(squirrel.Dollar)

	query := builder.
		Updt(messagesTable).
		Wr(squirrel.Eq{messagesEventIdCol: eventsId}).
		St(MessagesStatusCol, status)

	if _, err := query.ExcCtx(ctx, r.db); err != nil {
		return fmt.Errorf("%s : failed to execute query: %w", op, err)
	}

	return nil
}

func (r *EventRepository) GetEventById(ctx context.Context, eventId uuid.UUID) (*models.EventInfo, error) {
	op := "UserRepository.getEventById"

	builder := rep_utils.QBuilder.PHFormat(squirrel.Dollar)

	query := builder.
		Sel(messagesAllCol).
		Frm(messagesTable).
		Wr(squirrel.Eq{messagesEventIdCol: eventId})

	var eventInfo models.EventInfo
	err := query.QryRowCtx(ctx, r.db, &eventInfo)
	if err != nil {
		return nil, fmt.Errorf("%s : failed to execute query: %w", op, err)
	}

	return &eventInfo, nil
}

func (r *EventRepository) Create(ctx context.Context, info repositories_transfer.CreateEventInfo, tx *sql.Tx) error {
	op := "UserRepository.create"

	executor := rep_utils.GetExecutor(r.db, tx)
	builder := rep_utils.QBuilder.PHFormat(squirrel.Dollar)

	query := builder.
		Ins(messagesTable).
		Cols(messagesEventIdCol, messagesEventTypeCol, messagesPayloadCol).
		Vls(info.EventId, info.EventType, info.Payload)

	_, err := query.ExcCtx(ctx, executor)
	if err != nil {
		return fmt.Errorf("%s : failed to execute query: %w", op, err)
	}

	return nil
}
