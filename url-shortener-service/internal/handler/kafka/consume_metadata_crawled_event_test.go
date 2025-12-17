package kafka

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/require"
)

var testTime = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

func TestMetadataCrawled(t *testing.T) {
	tcs := map[string]struct {
		message       kafkago.Message
		wantRetriable bool
		wantErr       bool
	}{
		"success - valid message processed": {
			message: kafkago.Message{
				Topic:     "urlshortener.metadata.crawled.v1",
				Partition: 0,
				Offset:    1,
				Value: mustMarshal(model.Payload{
					EventID:    123,
					OccurredAt: testTime,
					Data: map[string]string{
						"short_code":   "abc123",
						"original_url": "https://example.com",
					},
					TraceID:       "12345678901234567890123456789012",
					SpanID:        "1234567890123456",
					CorrelationID: "corr-789",
				}),
			},
			wantRetriable: false,
			wantErr:       false,
		},

		"fail - invalid JSON payload": {
			message: kafkago.Message{
				Topic:     "urlshortener.metadata.crawled.v1",
				Partition: 0,
				Offset:    1,
				Value:     []byte("invalid json"),
			},
			wantRetriable: false,
			wantErr:       true,
		},

		"success - minimal payload": {
			message: kafkago.Message{
				Topic:     "urlshortener.metadata.crawled.v1",
				Partition: 0,
				Offset:    1,
				Value: mustMarshal(model.Payload{
					EventID:    456,
					OccurredAt: testTime,
					Data: map[string]string{
						"short_code": "xyz789",
					},
					TraceID:       "12345678901234567890123456789012",
					SpanID:        "1234567890123456",
					CorrelationID: "corr-123",
				}),
			},
			wantRetriable: false,
			wantErr:       false,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			handler := MetadataCrawled()
			err := handler.ConsumeMessage(ctx, tc.message)

			if tc.wantErr {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestNotifyMetadataCrawled(t *testing.T) {
	t.Run("success - notification placeholder", func(t *testing.T) {
		ctx := context.Background()
		err := notifyMetadataCrawled(ctx)
		require.NoError(t, err, "notifyMetadataCrawled should not return an error")
	})
}

// Helper function to marshal payload for tests
func mustMarshal(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

