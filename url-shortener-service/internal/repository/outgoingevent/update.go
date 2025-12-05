package outgoingevent

import (
	"context"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/infra/monitoring"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/orm"
	pkgerrors "github.com/pkg/errors"
)

// Update updates an existing outbox event using a dynamic whitelist
// based on which fields in the model have meaningful values.
// Only fields explicitly set by the caller will be updated.
func (i impl) Update(ctx context.Context, m model.OutgoingEvent, id int64) error {
	var err error
	ctx, span := monitoring.Start(ctx, "OutgoingEventRepository.Update")
	defer monitoring.End(span, &err)

	current, err := orm.FindOutgoingEvent(ctx, i.db, id)
	if err != nil {
		return pkgerrors.WithStack(err)
	}

	whitelist := []string{
		orm.OutgoingEventColumns.UpdatedAt,
	}

	if m.Status != "" {
		current.Status = m.Status.String()
		whitelist = append(whitelist, orm.OutgoingEventColumns.Status)
	}

	if m.LastError != "" {
		current.LastError = null.StringFrom(m.LastError)
		whitelist = append(whitelist, orm.OutgoingEventColumns.LastError)
	}

	if m.RetryCount > 0 {
		current.RetryCount = m.RetryCount
		whitelist = append(whitelist, orm.OutgoingEventColumns.RetryCount)
	}

	_, err = current.Update(ctx, i.db, boil.Whitelist(whitelist...))
	if err != nil {
		return pkgerrors.WithStack(err)
	}

	return nil
}
