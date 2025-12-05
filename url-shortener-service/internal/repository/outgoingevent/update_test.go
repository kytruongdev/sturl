package outgoingevent

import (
	"context"
	"database/sql"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestUpdate(t *testing.T) {
	tcs := map[string]struct {
		fixture string
		id      int64
		update  model.OutgoingEvent
		want    model.OutgoingEvent
		wantErr bool
	}{
		"success - update status only": {
			fixture: "testdata/outgoing_events.sql",
			id:      1,
			update: model.OutgoingEvent{
				Status: model.OutgoingEventStatusPublished,
			},
			want: model.OutgoingEvent{
				ID:     1,
				Topic:  model.TopicMetadataRequestedV1,
				Status: model.OutgoingEventStatusPublished,
			},
			wantErr: false,
		},

		"success - update status to failed with error message": {
			fixture: "testdata/outgoing_events.sql",
			id:      1,
			update: model.OutgoingEvent{
				Status:    model.OutgoingEventStatusFailed,
				LastError: "network timeout",
			},
			want: model.OutgoingEvent{
				ID:        1,
				Topic:     model.TopicMetadataRequestedV1,
				Status:    model.OutgoingEventStatusFailed,
				LastError: "network timeout",
			},
			wantErr: false,
		},

		"success - update retry count": {
			fixture: "testdata/outgoing_events.sql",
			id:      1,
			update: model.OutgoingEvent{
				RetryCount: 3,
				Status:     model.OutgoingEventStatusPending,
				LastError:  "retry attempt 3",
			},
			want: model.OutgoingEvent{
				ID:         1,
				Topic:      model.TopicMetadataRequestedV1,
				Status:     model.OutgoingEventStatusPending,
				RetryCount: 3,
				LastError:  "retry attempt 3",
			},
			wantErr: false,
		},

		"success - update to published": {
			fixture: "testdata/outgoing_events.sql",
			id:      1,
			update: model.OutgoingEvent{
				Status: model.OutgoingEventStatusPublished,
			},
			want: model.OutgoingEvent{
				ID:     1,
				Topic:  model.TopicMetadataRequestedV1,
				Status: model.OutgoingEventStatusPublished,
			},
			wantErr: false,
		},

		"fail - event not found": {
			fixture: "testdata/outgoing_events.sql",
			id:      99999,
			update: model.OutgoingEvent{
				Status: model.OutgoingEventStatusFailed,
			},
			wantErr: true,
		},

		"success - clear error on retry": {
			fixture: "testdata/outgoing_events.sql",
			id:      2, // Assuming event 2 exists with some error
			update: model.OutgoingEvent{
				Status:     model.OutgoingEventStatusPending,
				RetryCount: 1,
				LastError:  "previous error cleared, retrying",
			},
			want: model.OutgoingEvent{
				ID:         2,
				Status:     model.OutgoingEventStatusPending,
				RetryCount: 1,
				LastError:  "previous error cleared, retrying",
			},
			wantErr: false,
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

				err := repo.Update(ctx, tc.update, tc.id)

				if tc.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)

				// Verify the update by fetching all events and finding the updated one
				events, err := repo.GetByStatus(ctx, tc.want.Status.String(), 100)
				require.NoError(t, err)

				// Find the updated event by ID
				var found *model.OutgoingEvent
				for _, evt := range events {
					if evt.ID == tc.id {
						found = &evt
						break
					}
				}

				require.NotNil(t, found, "updated event not found with ID %d", tc.id)

				// Compare the updated event
				require.True(t,
					cmp.Equal(tc.want, *found,
						cmpopts.IgnoreFields(model.OutgoingEvent{},
							"CreatedAt", "UpdatedAt", "Payload", "CorrelationID", "TraceID", "SpanID", "Topic")),
					"diff: %v",
					cmp.Diff(tc.want, *found,
						cmpopts.IgnoreFields(model.OutgoingEvent{},
							"CreatedAt", "UpdatedAt", "Payload", "CorrelationID", "TraceID", "SpanID", "Topic")),
				)

				// Verify specific fields
				require.Equal(t, tc.want.Status, found.Status)
				if tc.update.LastError != "" {
					require.Equal(t, tc.want.LastError, found.LastError)
				}
				if tc.update.RetryCount > 0 {
					require.Equal(t, tc.want.RetryCount, found.RetryCount)
				}
			})
		})
	}
}

