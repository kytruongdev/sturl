package kafkaoutboxevent

import (
	"context"
	"encoding/json"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

// GetByStatus retrieves a kafka outbox event record by its status
func (i impl) GetByStatus(ctx context.Context, status string) ([]model.KafkaOutboxEvent, error) {
	items, err := orm.KafkaOutboxEvents(orm.KafkaOutboxEventWhere.Status.EQ(status), qm.OrderBy(orm.KafkaOutboxEventColumns.ID)).All(ctx, i.db)
	if err != nil {
		return nil, pkgerrors.WithStack(err)
	}

	var rs []model.KafkaOutboxEvent
	for _, item := range items {
		m, err := toKafkaOutboxEventModel(item)
		if err != nil {
			return nil, pkgerrors.WithStack(err)
		}

		rs = append(rs, m)
	}

	return rs, nil
}

func toKafkaOutboxEventModel(o *orm.KafkaOutboxEvent) (model.KafkaOutboxEvent, error) {
	m := model.KafkaOutboxEvent{
		ID:        o.ID,
		EventType: o.EventType,
		Status:    model.KafkaOutboxEventStatus(o.Status),
		CreatedAt: o.CreatedAt,
		UpdatedAt: o.UpdatedAt,
	}

	if err := json.Unmarshal(o.Payload, &m.Payload); err != nil {
		return model.KafkaOutboxEvent{}, pkgerrors.WithStack(err)
	}

	return m, nil
}
