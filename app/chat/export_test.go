package chat

import (
	"context"
	"testing"

	"github.com/expr-lang/expr"
	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
)

func TestExportExpandsGroupedMediaWhenFilteredMessageMatches(t *testing.T) {
	ctx := context.Background()
	filter, err := expr.Compile(`Message contains "#sample"`, expr.AsBool())
	require.NoError(t, err)

	grouped := []*tg.Message{
		testExportVideoMessage(10, 100, ""),
		testExportVideoMessage(11, 100, `#sample`),
		testExportVideoMessage(12, 100, ""),
	}
	calls := 0
	expander := newExportMessageExpander(ExportOptions{}, filter, func(context.Context, *tg.Message) ([]*tg.Message, error) {
		calls++
		return grouped, nil
	})

	messages, err := expander.Expand(ctx, grouped[1])
	require.NoError(t, err)
	require.Equal(t, []int{10, 11, 12}, exportMessageIDs(messages))

	messages, err = expander.Expand(ctx, grouped[1])
	require.NoError(t, err)
	require.Empty(t, messages)
	require.Equal(t, 1, calls)
}

func TestExportDoesNotExpandGroupedMediaWhenFilterDoesNotMatch(t *testing.T) {
	ctx := context.Background()
	filter, err := expr.Compile(`Message contains "#sample"`, expr.AsBool())
	require.NoError(t, err)

	expander := newExportMessageExpander(ExportOptions{}, filter, func(context.Context, *tg.Message) ([]*tg.Message, error) {
		t.Fatal("group resolver should not run when the filtered message does not match")
		return nil, nil
	})

	messages, err := expander.Expand(ctx, testExportVideoMessage(10, 100, "other"))
	require.NoError(t, err)
	require.Empty(t, messages)
}

func exportMessageIDs(messages []*Message) []int {
	ids := make([]int, 0, len(messages))
	for _, message := range messages {
		ids = append(ids, message.ID)
	}
	return ids
}

func testExportVideoMessage(id int, groupedID int64, text string) *tg.Message {
	msg := &tg.Message{
		ID:      id,
		Date:    1,
		Message: text,
	}
	msg.SetGroupedID(groupedID)
	msg.SetMedia(&tg.MessageMediaDocument{
		Document: &tg.Document{
			ID:       int64(id),
			MimeType: "video/mp4",
			Size:     1024,
			DCID:     1,
			Attributes: []tg.DocumentAttributeClass{
				&tg.DocumentAttributeFilename{FileName: "video.mp4"},
			},
		},
	})
	return msg
}
