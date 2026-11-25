package tmessage

import (
	"github.com/gotd/td/tg"
)

type Dialog struct {
	Peer        tg.InputPeerClass
	Messages    []int
	MediaCounts map[int]int
}

// MediaCount returns the number of downloadable items represented by a
// message. Most messages contain one item, while paid media can contain many.
func (d *Dialog) MediaCount(messageID int) int {
	if count, ok := d.MediaCounts[messageID]; ok {
		return count
	}
	return 1
}

type ParseSource func() ([]*Dialog, error)

func Parse(src ParseSource) ([]*Dialog, error) {
	return src()
}
