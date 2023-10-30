package dliter

import (
	"sync"
	"text/template"

	"github.com/gotd/td/telegram/peers"

	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/tmessage"
)

type Options struct {
	Pool             dcpool.Pool
	KV               kv.KV
	Template         string
	Include, Exclude []string
	Desc             bool
	Dialogs          [][]*tmessage.Dialog
}

type Iter struct {
	pool             dcpool.Pool
	dialogs          []*tmessage.Dialog
	include, exclude map[string]struct{}
	mu               sync.Mutex
	curi             int
	curj             int
	preSum           []int
	finished         map[int]struct{}
	template         *template.Template
	manager          *peers.Manager
	fingerprint      string
}

type fileTemplate struct {
	DialogID     int64
	MessageID    int
	MessageDate  int64
	FileName     string
	FileSize     string
	DownloadDate int64
}
