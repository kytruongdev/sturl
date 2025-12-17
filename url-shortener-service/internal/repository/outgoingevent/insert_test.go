package outgoingevent

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
		given   model.OutgoingEvent
		want    model.OutgoingEvent
		wantErr error
	}{
		"success": {
			given: model.OutgoingEvent{
				ID:     -10,
				Topic:  model.TopicMetadataRequestedV1,
				Status: model.OutgoingEventStatusPending,
				Payload: model.Payload{
					EventID: -10,
					Data: map[string]string{
						"short_code":   "abc123",
						"original_url": "https://google.com",
					},
				},
			},
			want: model.OutgoingEvent{
				ID:     -10,
				Topic:  model.TopicMetadataRequestedV1,
				Status: model.OutgoingEventStatusPending,
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
			fixture: "testdata/outgoing_events.sql",
			given: model.OutgoingEvent{
				ID:     1,
				Topic:  "evt.duplicate",
				Status: model.OutgoingEventStatusPending,
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
							cmpopts.IgnoreFields(model.OutgoingEvent{},
								"CreatedAt", "UpdatedAt", "Payload")),
						"diff: %v",
						cmp.Diff(tc.want, actual,
							cmpopts.IgnoreFields(model.OutgoingEvent{},
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
