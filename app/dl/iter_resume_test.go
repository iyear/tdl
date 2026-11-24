package dl

import (
	"context"
	"sync"
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/assert"

	"github.com/iyear/tdl/core/downloader"
	"github.com/iyear/tdl/pkg/tmessage"
)

// TestProcessSkipsFinishedWithoutAPI is the regression test for the --continue
// stall. When a message was already finished in a prior run, the iterator must
// skip it WITHOUT resolving it through the Telegram API.
//
// The nil pool/manager are intentional canaries: the old code path called
// manager.FromInputPeer (and then GetSingleMessage via the pool) before the
// finished check, which would dereference these nils and panic. The fast path
// returns before reaching them, so this test only passes with the fix.
func TestProcessSkipsFinishedWithoutAPI(t *testing.T) {
	it := &iter{
		mu:   &sync.Mutex{},
		opts: Options{Group: false},
		dialogs: []*tmessage.Dialog{{
			Peer:     &tg.InputPeerChat{ChatID: 1},
			Messages: []int{100},
		}},
		finished:   map[int]struct{}{0: {}}, // logicalPos 0 already finished
		logicalPos: 0,
		elem:       make(chan downloader.Elem, 10),
	}

	ret, skip := it.process(context.Background())

	// Skipped, no element produced, logicalPos advanced past the finished slot.
	assert.False(t, ret)
	assert.True(t, skip)
	assert.Equal(t, 1, it.logicalPos)
	assert.Equal(t, 1, it.dialogIndex, "physical position must still advance")
}

// TestProcessGroupModeDoesNotFastSkip proves the fast path is gated on !Group.
// In group mode a finished logicalPos must NOT be skipped here, because album
// size (and thus logicalPos advancement) is only known after resolving. The
// iterator falls through to resolve the message, dereferencing the nil manager
// and panicking — which is exactly the behavior we want to preserve.
func TestProcessGroupModeDoesNotFastSkip(t *testing.T) {
	it := &iter{
		mu:   &sync.Mutex{},
		opts: Options{Group: true},
		dialogs: []*tmessage.Dialog{{
			Peer:     &tg.InputPeerChat{ChatID: 1},
			Messages: []int{100},
		}},
		finished:   map[int]struct{}{0: {}},
		logicalPos: 0,
		elem:       make(chan downloader.Elem, 10),
	}

	assert.Panics(t, func() { _, _ = it.process(context.Background()) },
		"group mode must fall through to resolve despite finished hit")
}
