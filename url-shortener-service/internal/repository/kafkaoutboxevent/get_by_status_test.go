package kafkaoutboxevent

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
		want    []model.KafkaOutboxEvent
		wantErr error
	}{
		"success - return PENDING events": {
			fixture: "testdata/kafka_outbox_events.sql",
			status:  "PENDING",
			want: []model.KafkaOutboxEvent{
				{
					ID:        1,
					EventType: "evt.pending.1",
					Status:    model.KafkaOutboxEventStatusPending,
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
					ID:        2,
					EventType: "evt.pending.2",
					Status:    model.KafkaOutboxEventStatusPending,
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
					ID:        3,
					EventType: "evt.pending.3",
					Status:    model.KafkaOutboxEventStatusPending,
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
			fixture: "testdata/kafka_outbox_events.sql",
			status:  "FAILED",
			want:    nil,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {

			testutil.WithTxDB(t, func(tx *sql.Tx) {
				testutil.LoadSQLFile(t, tx, tc.fixture)
				ctx := context.Background()

				repo := New(tx)

				got, err := repo.GetByStatus(ctx, tc.status)

				if tc.wantErr != nil {
					require.EqualError(t, err, tc.wantErr.Error())
					return
				}

				require.NoError(t, err)

				require.True(t,
					cmp.Equal(tc.want, got,
						cmpopts.IgnoreFields(model.KafkaOutboxEvent{},
							"CreatedAt", "UpdatedAt", "Payload.OccurredAt")),
					"diff: %v",
					cmp.Diff(tc.want, got,
						cmpopts.IgnoreFields(model.KafkaOutboxEvent{},
							"CreatedAt", "UpdatedAt", "Payload.OccurredAt")),
				)
			})
		})
	}
}
