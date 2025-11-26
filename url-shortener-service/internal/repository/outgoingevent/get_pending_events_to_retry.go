package outgoingevent

import (
	"context"

	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

// GetPendingEventsToRetry retrieves PENDING kafka outbox event records to retry
func (i impl) GetPendingEventsToRetry(ctx context.Context, status string, limit int) ([]model.OutgoingEvent, error) {
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
