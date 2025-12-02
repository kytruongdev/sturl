package outgoingevent

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

// Insert creates a new outbox event record in the database.
func (i impl) Insert(ctx context.Context, m model.OutgoingEvent) (model.OutgoingEvent, error) {
	var err error
	ctx, span := monitoring.Start(ctx, "OutgoingEventRepository.Insert")
	defer monitoring.End(span, &err)

	o := orm.OutgoingEvent{
		ID:            m.ID,
		Topic:         m.Topic.String(),
		Status:        m.Status.String(),
		CorrelationID: m.CorrelationID,
		TraceID:       m.TraceID,
		SpanID:        m.SpanID,
	}

	o.Payload, err = m.MarshalPayload()
	if err != nil {
		return model.OutgoingEvent{}, pkgerrors.WithStack(err)
	}

	if err = o.Insert(ctx, i.db, boil.Infer()); err != nil {
		return model.OutgoingEvent{}, pkgerrors.WithStack(err)
	}

	m.CreatedAt = o.CreatedAt
	m.UpdatedAt = o.UpdatedAt

	return m, nil
}
