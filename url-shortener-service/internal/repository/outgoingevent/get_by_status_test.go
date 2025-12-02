package outgoingevent

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestGetByStatus(t *testing.T) {
	tcs := map[string]struct {
		fixture string
		status  string
		limit   int
		want    []model.OutgoingEvent
		wantErr error
	}{
		"success - return PENDING events": {
			fixture: "testdata/outgoing_events.sql",
			status:  "PENDING",
			limit:   100,
			want: []model.OutgoingEvent{
				{
					ID:            1,
					Topic:         "evt.pending.1",
					Status:        model.OutgoingEventStatusPending,
					CorrelationID: "111222",
					TraceID:       "bbb",
					SpanID:        "ccc",
					Payload: model.Payload{
						EventID:    1,
						OccurredAt: time.Now(),
						Data: map[string]string{
							"short_code":   "1",
							"original_url": "foo.bar/1",
						},
					},
				},
				{
					ID:            2,
					Topic:         "evt.pending.2",
					Status:        model.OutgoingEventStatusPending,
					CorrelationID: "qwerty",
					TraceID:       "999",
					SpanID:        "888",
					Payload: model.Payload{
						EventID:    2,
						OccurredAt: time.Now(),
						Data: map[string]string{
							"short_code":   "2",
							"original_url": "foo.bar/2",
						},
					},
				},
				{
					ID:            3,
					Topic:         "evt.pending.3",
					Status:        model.OutgoingEventStatusPending,
					CorrelationID: "a1b2c3",
					TraceID:       "xxx",
					SpanID:        "yyy",
					Payload: model.Payload{
						EventID:    3,
						OccurredAt: time.Now(),
						Data: map[string]string{
							"short_code":   "3",
							"original_url": "foo.bar/3",
						},
					},
				},
			},
		},
		"success - empty result": {
			fixture: "testdata/outgoing_events.sql",
			status:  "FAILED",
			limit:   100,
			want:    nil,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {

			testutil.WithTxDB(t, func(tx *sql.Tx) {
				testutil.LoadSQLFile(t, tx, tc.fixture)
				ctx := context.Background()

				got, err := New(tx).GetByStatus(ctx, tc.status, tc.limit)

				if tc.wantErr != nil {
					require.EqualError(t, err, tc.wantErr.Error())
					return
				}

				require.NoError(t, err)

				require.True(t,
					cmp.Equal(tc.want, got,
						cmpopts.IgnoreFields(model.OutgoingEvent{},
							"CreatedAt", "UpdatedAt", "Payload.OccurredAt")),
					"diff: %v",
					cmp.Diff(tc.want, got,
						cmpopts.IgnoreFields(model.OutgoingEvent{},
							"CreatedAt", "UpdatedAt", "Payload.OccurredAt")),
				)
			})
		})
	}
}
