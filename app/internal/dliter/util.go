package dliter

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"sort"

	"github.com/iyear/tdl/pkg/utils"
)

func sortDialogs(dialogs []*Dialog, desc bool) {
	sort.Slice(dialogs, func(i, j int) bool {
		return utils.Telegram.GetInputPeerID(dialogs[i].Peer) <
			utils.Telegram.GetInputPeerID(dialogs[j].Peer) // increasing order
	})

	for _, m := range dialogs {
		sort.Slice(m.Messages, func(i, j int) bool {
			if desc {
				return m.Messages[i] > m.Messages[j]
			}
			return m.Messages[i] < m.Messages[j]
		})
	}
}

func fingerprint(dialogs []*Dialog) string {
	endian := binary.BigEndian
	buf, b := &bytes.Buffer{}, make([]byte, 8)
	for _, m := range dialogs {
		endian.PutUint64(b, uint64(utils.Telegram.GetInputPeerID(m.Peer)))
		buf.Write(b)
		for _, msg := range m.Messages {
			endian.PutUint64(b, uint64(msg))
			buf.Write(b)
		}
	}

	return fmt.Sprintf("%x", sha256.Sum256(buf.Bytes()))
}

func filterMap(data []string, keyFn func(key string) string) map[string]struct{} {
	m := make(map[string]struct{})
	for _, v := range data {
		m[keyFn(v)] = struct{}{}
	}
	return m
}

func collectDialogs(dialogs [][]*Dialog) []*Dialog {
	res := make([]*Dialog, 0)
	for _, d := range dialogs {
		if len(d) == 0 {
			continue
		}
		res = append(res, d...)
	}
	return res
}

// preSum of dialogs
func preSum(dialogs []*Dialog) []int {
	sum := make([]int, len(dialogs)+1)
	for i, m := range dialogs {
		sum[i+1] = sum[i] + len(m.Messages)
	}
	return sum
}
