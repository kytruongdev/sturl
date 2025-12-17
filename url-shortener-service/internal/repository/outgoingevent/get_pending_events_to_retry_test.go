package outgoingevent

import (
	"context"
	"database/sql"
	"testing"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/pkg/testutil"
	"github.com/stretchr/testify/require"
)

func TestGetPendingEventsToRetry(t *testing.T) {
	tcs := map[string]struct {
		fixture   string
		status    string
		limit     int
		wantCount int
		wantErr   bool
	}{
		"success - get pending events with limit": {
			fixture:   "testdata/outgoing_events.sql",
			status:    model.OutgoingEventStatusPending.String(),
			limit:     10,
			wantCount: 3, // There are 3 PENDING events in fixture
			wantErr:   false,
		},

		"success - get pending events with small limit": {
			fixture:   "testdata/outgoing_events.sql",
			status:    model.OutgoingEventStatusPending.String(),
			limit:     1,
			wantCount: 1,
			wantErr:   false,
		},

		"success - get failed events": {
			fixture:   "testdata/outgoing_events.sql",
			status:    model.OutgoingEventStatusFailed.String(),
			limit:     10,
			wantCount: 0, // Based on fixture, adjust if needed
			wantErr:   false,
		},

		"success - no events found": {
			fixture:   "testdata/outgoing_events.sql",
			status:    "NON_EXISTENT_STATUS",
			limit:     10,
			wantCount: 0,
			wantErr:   false,
		},

		"success - limit zero returns nothing": {
			fixture:   "testdata/outgoing_events.sql",
			status:    model.OutgoingEventStatusPending.String(),
			limit:     0,
			wantCount: 0,
			wantErr:   false,
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

				events, err := repo.GetPendingEventsToRetry(ctx, tc.status, tc.limit)

				if tc.wantErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				require.Len(t, events, tc.wantCount, "expected %d events, got %d", tc.wantCount, len(events))

				// Verify events are ordered by ID
				if len(events) > 1 {
					for i := 1; i < len(events); i++ {
						require.Greater(t, events[i].ID, events[i-1].ID,
							"events should be ordered by ID ascending")
					}
				}

				// Verify all events have the correct status
				for _, evt := range events {
					require.Equal(t, tc.status, evt.Status.String(),
						"event %d should have status %s", evt.ID, tc.status)
				}

				// Verify limit is respected
				require.LessOrEqual(t, len(events), tc.limit,
					"result count should not exceed limit")
			})
		})
	}
}

func TestGetPendingEventsToRetry_WithMultipleEvents(t *testing.T) {
	t.Run("success - multiple pending events ordered by ID", func(t *testing.T) {
		testutil.WithTxDB(t, func(tx *sql.Tx) {
			ctx := context.Background()
			repo := New(tx)

			// Insert multiple pending events
			event1 := model.OutgoingEvent{
				ID:     -100,
				Topic:  model.TopicMetadataRequestedV1,
				Status: model.OutgoingEventStatusPending,
				Payload: model.Payload{
					EventID: -100,
					Data: map[string]string{
						"test": "event1",
					},
				},
			}
			event2 := model.OutgoingEvent{
				ID:     -101,
				Topic:  model.TopicMetadataRequestedV1,
				Status: model.OutgoingEventStatusPending,
				Payload: model.Payload{
					EventID: -101,
					Data: map[string]string{
						"test": "event2",
					},
				},
			}
			event3 := model.OutgoingEvent{
				ID:     -102,
				Topic:  model.TopicMetadataRequestedV1,
				Status: model.OutgoingEventStatusPending,
				Payload: model.Payload{
					EventID: -102,
					Data: map[string]string{
						"test": "event3",
					},
				},
			}

			_, err := repo.Insert(ctx, event1)
			require.NoError(t, err)
			_, err = repo.Insert(ctx, event2)
			require.NoError(t, err)
			_, err = repo.Insert(ctx, event3)
			require.NoError(t, err)

			// Get pending events with limit
			events, err := repo.GetPendingEventsToRetry(ctx, model.OutgoingEventStatusPending.String(), 2)
			require.NoError(t, err)
			require.Len(t, events, 2, "should return exactly 2 events due to limit")

			// Verify ordering - should get the first 2 by ID (ascending)
			require.Equal(t, int64(-102), events[0].ID, "first event should be -102")
			require.Equal(t, int64(-101), events[1].ID, "second event should be -101")

			// Get all pending events
			allEvents, err := repo.GetPendingEventsToRetry(ctx, model.OutgoingEventStatusPending.String(), 100)
			require.NoError(t, err)
			require.Len(t, allEvents, 3, "should return all 3 pending events")

			// Verify all have correct status
			for _, evt := range allEvents {
				require.Equal(t, model.OutgoingEventStatusPending.String(), evt.Status.String())
			}
		})
	})
}

