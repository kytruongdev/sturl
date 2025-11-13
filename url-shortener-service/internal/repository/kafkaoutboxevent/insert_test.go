package kafkaoutboxevent

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestInsert(t *testing.T) {
	tcs := map[string]struct {
		fixture string
		given   model.KafkaOutboxEvent
		want    model.KafkaOutboxEvent
		wantErr error
	}{
		"success": {
			given: model.KafkaOutboxEvent{
				ID:        -10,
				EventType: "urlshortener.metadata.requested.v1",
				Status:    model.KafkaOutboxEventStatusPending,
				Payload: model.Payload{
					EventID: -10,
					Data: map[string]string{
						"short_code":   "abc123",
						"original_url": "https://google.com",
					},
				},
			},
			want: model.KafkaOutboxEvent{
				ID:        -10,
				EventType: "urlshortener.metadata.requested.v1",
				Status:    model.KafkaOutboxEventStatusPending,
				Payload: model.Payload{
					EventID: -10,
					Data: map[string]string{
						"short_code":   "abc123",
						"original_url": "https://google.com",
					},
				},
			},
		},
		"fail - duplicate primary key": {
			fixture: "testdata/kafka_outbox_events.sql",
			given: model.KafkaOutboxEvent{
				ID:        1,
				EventType: "evt.duplicate",
				Status:    model.KafkaOutboxEventStatusPending,
				Payload: model.Payload{
					EventID: 1,
				},
			},
			wantErr: errors.New(`duplicate key value violates unique constraint`),
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			testutil.WithTxDB(t, func(tx *sql.Tx) {
				if tc.fixture != "" {
					testutil.LoadSQLFile(t, tx, tc.fixture)
				}

				ctx := context.Background()
				repo := New(tx)

				actual, err := repo.Insert(ctx, tc.given)

				if tc.wantErr != nil {
					require.Error(t, err)
					require.Contains(t, err.Error(), tc.wantErr.Error())
				} else {
					require.NoError(t, err)
					require.True(t,
						cmp.Equal(tc.want, actual,
							cmpopts.IgnoreFields(model.KafkaOutboxEvent{},
								"CreatedAt", "UpdatedAt", "Payload")),
						"diff: %v",
						cmp.Diff(tc.want, actual,
							cmpopts.IgnoreFields(model.KafkaOutboxEvent{},
								"CreatedAt", "UpdatedAt", "Payload")),
					)

					var unmarshaled model.Payload
					b, err := actual.MarshalPayload()
					require.NoError(t, err)
					err = json.Unmarshal(b, &unmarshaled)
					require.NoError(t, err)

					require.Equal(t, tc.given.Payload.EventID, unmarshaled.EventID)
				}
			})
		})
	}
}
