package up

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/expr-lang/expr/vm"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-faster/errors"
	"github.com/go-viper/mapstructure/v2"
	"github.com/gotd/td/telegram/peers"
	"go.uber.org/zap"

	"github.com/iyear/tdl/core/logctx"
	"github.com/iyear/tdl/core/uploader"
	"github.com/iyear/tdl/core/util/mediautil"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/texpr"
)

type Env struct {
	File      string `comment:"File path"`
	Thumb     string `comment:"Thumbnail path"`
	Filename  string `comment:"Filename"`
	Extension string `comment:"File extension"`
	Mime      string `comment:"File mime type"`
}

func exprEnv(ctx context.Context, file *File) Env {
	if file == nil {
		return Env{}
	}

	var extension = filepath.Ext(file.File)
	var filename = strings.TrimSuffix(filepath.Base(file.File), extension)
	var mime, err = mimetype.DetectFile(file.File)
	if err != nil {
		mime = &mimetype.MIME{}
		logctx.From(ctx).Error("detect file mime", zap.Error(err))
	}
	return Env{
		File:      file.File,
		Thumb:     file.Thumb,
		Filename:  filename,
		Extension: extension,
		Mime:      mime.String(),
	}
}

type File struct {
	File  string
	Thumb string
}

type dest struct {
	Peer   string
	Thread int
}

type iter struct {
	files   []*File
	to      *vm.Program
	chat    string
	topic   int
	photo   bool
	remove  bool
	delay   time.Duration
	manager *peers.Manager

	cur  int
	err  error
	file uploader.Elem
}

func newIter(files []*File, to *vm.Program, chat string, topic int, photo, remove bool, delay time.Duration, manager *peers.Manager) *iter {
	return &iter{
		files:   files,
		to:      to,
		chat:    chat,
		topic:   topic,
		photo:   photo,
		remove:  remove,
		delay:   delay,
		manager: manager,

		cur:  0,
		err:  nil,
		file: nil,
	}
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		i.err = ctx.Err()
		return false
	default:
	}

	if i.cur >= len(i.files) || i.err != nil {
		return false
	}

	// if delay is set, sleep for a while for each iteration
	if i.delay > 0 && i.cur > 0 { // skip first delay
		time.Sleep(i.delay)
	}

	cur := i.files[i.cur]
	i.cur++

	f, err := os.Open(cur.File)
	if err != nil {
		i.err = errors.Wrap(err, "open file")
		return false
	}

	var (
		to     peers.Peer
		thread int
	)
	if i.chat != "" {
		to, i.err = i.resolvePeer(ctx, i.chat)
		thread = i.topic
		if i.err != nil {
			return false
		}
	} else {
		// message routing
		result, err := texpr.Run(i.to, exprEnv(ctx, cur))
		if err != nil {
			i.err = errors.Wrap(err, "message routing")
			return false
		}

		switch r := result.(type) {
		case string:
			// pure chat, no reply to, which is a compatible with old version
			// and a convenient way to send message to self
			to, err = i.resolvePeer(ctx, r)
		case map[string]interface{}:
			// chat with reply to topic or message
			var d dest

			if err = mapstructure.WeakDecode(r, &d); err != nil {
				i.err = errors.Wrapf(err, "decode dest: %v", result)
				return false
			}

			to, err = i.resolvePeer(ctx, d.Peer)
			thread = d.Thread
		default:
			i.err = errors.Errorf("message router must return string or dest: %T", result)
			return false
		}

		if err != nil {
			i.err = err
			return false
		}
	}

	var thumb *uploaderFile = nil
	// has thumbnail
	if cur.Thumb != "" {
		tMime, err := mimetype.DetectFile(cur.Thumb)
		if err != nil || !mediautil.IsImage(tMime.String()) { // TODO(iyear): jpg only
			i.err = errors.Wrapf(err, "invalid thumbnail file: %v", cur.Thumb)
			return false
		}
		thumbFile, err := os.Open(cur.Thumb)
		if err != nil {
			i.err = errors.Wrap(err, "open thumbnail file")
			return false
		}

		thumb = &uploaderFile{File: thumbFile, size: 0}
	}

	stat, err := f.Stat()
	if err != nil {
		i.err = errors.Wrap(err, "stat file")
		return false
	}

	i.file = &iterElem{
		file:   &uploaderFile{File: f, size: stat.Size()},
		thumb:  thumb,
		to:     to,
		thread: thread,

		asPhoto: i.photo,
		remove:  i.remove,
	}

	return true
}

func (i *iter) resolvePeer(ctx context.Context, peer string) (peers.Peer, error) {
	if peer == "" { // self
		return i.manager.Self(ctx)
	}

	return tutil.GetInputPeer(ctx, i.manager, peer)
}

func (i *iter) Value() uploader.Elem {
	return i.file
}

func (i *iter) Err() error {
	return i.err
}
