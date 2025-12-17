package outgoingevent

import (
	"context"
	"encoding/json"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

// GetByStatus retrieves a kafka outbox event records by their status
func (i impl) GetByStatus(ctx context.Context, status string, limit int) ([]model.OutgoingEvent, error) {
	var err error
	ctx, span := monitoring.Start(ctx, "OutgoingEventRepository.GetByStatus")
	defer monitoring.End(span, &err)

	items, err := orm.OutgoingEvents(
		orm.OutgoingEventWhere.Status.EQ(status),
		qm.OrderBy(orm.OutgoingEventColumns.ID),
		qm.Limit(limit)).All(ctx, i.db)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	var rs []model.OutgoingEvent
	for _, item := range items {
		m, err := toOutgoingEventModel(item)
		if err != nil {
			return nil, pkgerrors.WithStack(err)
		}

		rs = append(rs, m)
	}

	return rs, nil
}

func toOutgoingEventModel(o *orm.OutgoingEvent) (model.OutgoingEvent, error) {
	m := model.OutgoingEvent{
		ID:            o.ID,
		Topic:         model.Topic(o.Topic),
		RetryCount:    o.RetryCount,
		LastError:     o.LastError.String,
		CorrelationID: o.CorrelationID,
		TraceID:       o.TraceID,
		SpanID:        o.SpanID,
		Status:        model.OutgoingEventStatus(o.Status),
		CreatedAt:     o.CreatedAt,
		UpdatedAt:     o.UpdatedAt,
	}

	if err := json.Unmarshal(o.Payload, &m.Payload); err != nil {
		return model.OutgoingEvent{}, pkgerrors.WithStack(err)
	}

	return m, nil
}
