package tmessage

import (
	"github.com/gotd/td/tg"
)

type MessageInfo struct {
	ID   int
	File string
}

type Dialog struct {
	Peer     tg.InputPeerClass
	Messages []int
	FileInfo map[int]string
}

type ParseSource func() ([]*Dialog, error)

func Parse(src ParseSource) ([]*Dialog, error) {
	return src()
}
