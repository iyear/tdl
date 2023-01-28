package dliter

import (
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/kv"
	"sync"
	"text/template"
)

type Options struct {
	Pool             dcpool.Pool
	KV               kv.KV
	Template         string
	Include, Exclude []string
	Dialogs          [][]*Dialog
}

type Iter struct {
	pool             dcpool.Pool
	dialogs          []*Dialog
	include, exclude map[string]struct{}
	mu               sync.Mutex
	curi             int
	curj             int
	finished         map[int]struct{}
	template         *template.Template
	manager          *peers.Manager
	fingerprint      string
}

type Dialog struct {
	Peer     tg.InputPeerClass
	Messages []int
}

type fileTemplate struct {
	DialogID     int64
	MessageID    int
	MessageDate  int64
	FileName     string
	FileSize     string
	DownloadDate int64
}
