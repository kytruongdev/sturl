package kafkaoutboxevent

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

// Insert creates a new outbox event record in the database.
func (i impl) Insert(ctx context.Context, m model.KafkaOutboxEvent) (model.KafkaOutboxEvent, error) {
	var err error
	ctx, span := monitoring.Start(ctx, "OutboxEventRepository.Insert")
	defer monitoring.End(span, &err)

	o := orm.KafkaOutboxEvent{
		ID:        m.ID,
		EventType: m.EventType,
		Status:    m.Status.String(),
	}

	o.Payload, err = m.MarshalPayload()
	if err != nil {
		return model.KafkaOutboxEvent{}, pkgerrors.WithStack(err)
	}

	if err = o.Insert(ctx, i.db, boil.Infer()); err != nil {
		return model.KafkaOutboxEvent{}, pkgerrors.WithStack(err)
	}

	m.CreatedAt = o.CreatedAt
	m.UpdatedAt = o.UpdatedAt

	return m, nil
}
