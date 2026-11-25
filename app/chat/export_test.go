package chat

import (
	"context"
	"math"
	"testing"

	"github.com/gotd/td/telegram/query/messages"
	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
)

func TestExportHistoryContinuesAfterShortPage(t *testing.T) {
	ctx := t.Context()
	opts := ExportOptions{Type: ExportTypeTime, Input: []int{100, 200}}
	requests := []messages.Request{
		{OffsetDate: 201, Limit: 100},
		{OffsetID: 9, Limit: 100},
		{OffsetID: 7, Limit: 100},
		{OffsetID: 5, Limit: 100},
	}
	pages := []tg.MessagesMessagesClass{
		&tg.MessagesMessagesSlice{Messages: []tg.MessageClass{
			&tg.Message{ID: 9, Date: 200},
			&tg.Message{ID: 10, Date: 200},
		}},
		&tg.MessagesMessagesSlice{Messages: []tg.MessageClass{
			&tg.MessageEmpty{ID: 7},
			&tg.Message{ID: 8, Date: 200}, // Same timestamp as the first page's cursor.
		}},
		&tg.MessagesMessagesSlice{Messages: []tg.MessageClass{
			&tg.Message{ID: 5, Date: 199},
		}},
		&tg.MessagesMessagesSlice{},
	}

	var calls int
	q := messages.QueryFunc(func(_ context.Context, req messages.Request) (tg.MessagesMessagesClass, error) {
		require.Less(t, calls, len(requests))
		require.Equal(t, requests[calls], req)
		page := pages[calls]
		calls++
		return page, nil
	})

	var ids []int
	for msg, err := range exportHistory(ctx, q, opts) {
		require.NoError(t, err)
		ids = append(ids, msg.GetID())
	}
	require.Equal(t, []int{10, 9, 8, 5}, ids)
	require.Equal(t, len(requests), calls)
}

func TestExportHistoryInitialOffset(t *testing.T) {
	for _, tt := range []struct {
		name string
		typ  ExportType
		end  int
		want messages.Request
	}{
		{"unbounded time", ExportTypeTime, math.MaxInt, messages.Request{Limit: 100}},
		{"unbounded ID", ExportTypeId, math.MaxInt, messages.Request{Limit: 100}},
		{"inclusive ID", ExportTypeId, 42, messages.Request{OffsetID: 43, Limit: 100}},
		{"maximum wire ID", ExportTypeId, math.MaxInt32, messages.Request{Limit: 100}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var calls int
			q := messages.QueryFunc(func(_ context.Context, req messages.Request) (tg.MessagesMessagesClass, error) {
				calls++
				require.Equal(t, tt.want, req)
				return &tg.MessagesMessages{}, nil
			})

			opts := ExportOptions{Type: tt.typ, Input: []int{0, tt.end}}
			for _, err := range exportHistory(t.Context(), q, opts) {
				t.Fatalf("empty response yielded a result: %v", err)
			}
			require.Equal(t, 1, calls)
		})
	}
}
