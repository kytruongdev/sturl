package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	shortUrlCtrl "github.com/kytruongdev/sturl/url-shortener-service/internal/controller/shorturl"
	"github.com/kytruongdev/sturl/url-shortener-service/internal/model"
	kafkago "github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMetadataRequested(t *testing.T) {
	tcs := map[string]struct {
		message                   kafkago.Message
		mockCrawlMetadataResponse model.UrlMetadata
		mockCrawlMetadataErr      error
		wantErr                   bool
	}{
		"success - valid message processed": {
			message: kafkago.Message{
				Topic:     "urlshortener.metadata.requested.v1",
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
			mockCrawlMetadataResponse: model.UrlMetadata{
				Title:       "Example",
				Description: "Example site",
			},
			mockCrawlMetadataErr: nil,
			wantErr:              false,
		},

		"fail - invalid JSON payload": {
			message: kafkago.Message{
				Topic:     "urlshortener.metadata.requested.v1",
				Partition: 0,
				Offset:    1,
				Value:     []byte("invalid json"),
			},
			wantErr: true,
		},

		"fail - empty short_code": {
			message: kafkago.Message{
				Topic:     "urlshortener.metadata.requested.v1",
				Partition: 0,
				Offset:    1,
				Value: mustMarshal(model.Payload{
					EventID:    123,
					OccurredAt: testTime,
					Data: map[string]string{
						"short_code":   "",
						"original_url": "https://example.com",
					},
					TraceID:       "12345678901234567890123456789012",
					SpanID:        "1234567890123456",
					CorrelationID: "corr-789",
				}),
			},
			wantErr: true,
		},

		"fail - crawl metadata returns error": {
			message: kafkago.Message{
				Topic:     "urlshortener.metadata.requested.v1",
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
			mockCrawlMetadataErr: errors.New("network timeout"),
			wantErr:              true,
		},
	}

	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			// Mock controller
			mockCtrl := new(shortUrlCtrl.MockController)
			if tc.mockCrawlMetadataErr != nil || !tc.wantErr {
				var payload model.Payload
				_ = json.Unmarshal(tc.message.Value, &payload)
				shortCode := payload.Data["short_code"]

				if shortCode != "" {
					mockCtrl.On("CrawlURLMetadata", mock.Anything, shortCode).
						Return(tc.mockCrawlMetadataResponse, tc.mockCrawlMetadataErr)
				}
			}

			handler := MetadataRequested(mockCtrl)
			err := handler.ConsumeMessage(ctx, tc.message)

			if tc.wantErr {
				require.NotNil(t, err)
			} else {
				require.Nil(t, err)
			}
		})
	}
}

