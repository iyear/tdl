package up

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/expr-lang/expr/vm"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-faster/errors"
	"github.com/go-viper/mapstructure/v2"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/peers"

	"github.com/iyear/tdl/core/uploader"
	"github.com/iyear/tdl/core/util/mediautil"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/texpr"
)

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
	caption *vm.Program
	chat    string
	topic   int
	photo   bool
	remove  bool
	group   bool
	delay   time.Duration
	manager *peers.Manager

	cur     int
	err     error
	file    uploader.Elem
	pending *iterElem // first file of the next group
}

func newIter(files []*File, to, caption *vm.Program, chat string, topic int, photo, remove, group bool, delay time.Duration, manager *peers.Manager) *iter {
	return &iter{
		files:   files,
		to:      to,
		caption: caption,
		chat:    chat,
		topic:   topic,
		photo:   photo,
		remove:  remove,
		group:   group,
		delay:   delay,
		manager: manager,

		cur:     0,
		err:     nil,
		file:    nil,
		pending: nil,
	}
}

func (i *iter) Next(ctx context.Context) bool {
	if err := ctx.Err(); err != nil {
		if i.pending != nil {
			_ = i.pending.close()
			i.pending = nil
		}
		i.err = err
		return false
	}

	if i.err != nil || (i.cur >= len(i.files) && i.pending == nil) {
		return false
	}

	// if delay is set, sleep for a while for each iteration
	if i.delay > 0 && i.cur > 0 { // skip first delay
		time.Sleep(i.delay)
	}

	if i.group {
		return i.nextGroup(ctx)
	}

	cur := i.files[i.cur]
	i.cur++

	file, err := i.next(ctx, cur)
	if err != nil {
		i.err = err
		return false
	}

	i.file = file
	return true
}

// nextGroup groups consecutive compatible files for the same destination.
func (i *iter) nextGroup(ctx context.Context) bool {
	elems := make([]*iterElem, 0, uploader.MaxGroupSize)

	if i.pending != nil {
		elems = append(elems, i.pending)
		i.pending = nil
	}

	for len(elems) < uploader.MaxGroupSize && i.cur < len(i.files) {
		cur := i.files[i.cur]
		i.cur++

		elem, err := i.next(ctx, cur)
		if err != nil {
			for _, e := range elems {
				_ = e.close()
			}
			i.err = err
			return false
		}

		if len(elems) > 0 && !canGroup(elems[0], elem) {
			i.pending = elem
			break
		}

		elems = append(elems, elem)
	}

	for _, elem := range elems[1:] {
		elem.caption = &entity.Builder{}
	}

	i.file = &groupElem{elems: elems}
	return true
}

func canGroup(first, next *iterElem) bool {
	return first.to.TDLibPeerID() == next.to.TDLibPeerID() &&
		first.thread == next.thread &&
		first.media.CanGroupWith(next.media)
}

func (i *iter) next(ctx context.Context, cur *File) (_ *iterElem, rerr error) {
	file, err := i.resolveFile(cur.File)
	if err != nil {
		return nil, errors.Wrap(err, "resolve file")
	}
	defer func() {
		if rerr != nil {
			_ = file.Close()
		}
	}()

	media, err := uploader.InspectMedia(file, i.photo)
	if err != nil {
		return nil, errors.Wrapf(err, "inspect %s", cur.File)
	}
	env := exprEnv(cur)
	env.MIME = media.MIME()

	to, thread, err := i.resolveDest(ctx, env)
	if err != nil {
		return nil, errors.Wrap(err, "resolve destination")
	}

	caption, err := i.resolveCaption(env)
	if err != nil {
		return nil, errors.Wrap(err, "resolve caption")
	}

	thumb, err := i.resolveThumb(cur.Thumb)
	if err != nil {
		return nil, errors.Wrap(err, "resolve thumbnail")
	}

	return &iterElem{
		file:    file,
		thumb:   thumb,
		to:      to,
		caption: caption,
		thread:  thread,
		media:   media,

		asPhoto: i.photo,
		remove:  i.remove,
		grouped: i.group,
	}, nil
}

func (i *iter) resolveFile(path string) (*uploaderFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "open file")
	}

	stat, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, errors.Wrap(err, "stat file")
	}

	return &uploaderFile{
		File: f,
		size: stat.Size(),
	}, nil
}

func (i *iter) resolveDest(ctx context.Context, env Env) (peers.Peer, int, error) {
	if i.chat != "" { // compatible with old version
		to, err := i.resolvePeer(ctx, i.chat)
		if err != nil {
			return nil, 0, errors.Wrap(err, "resolve chat")
		}

		return to, i.topic, nil
	}

	// message routing
	result, err := texpr.Run(i.to, env)
	if err != nil {
		return nil, 0, errors.Wrap(err, "parse expression")
	}

	var (
		to     peers.Peer
		thread int
	)

	switch r := result.(type) {
	case string:
		// pure chat, no reply to, which is a compatible with old version
		// and a convenient way to send message to self
		to, err = i.resolvePeer(ctx, r)
	case map[string]interface{}:
		// chat with reply to topic or message
		var d dest

		if err = mapstructure.WeakDecode(r, &d); err != nil {
			return nil, 0, errors.Wrapf(err, "decode dest: %v", result)
		}

		to, err = i.resolvePeer(ctx, d.Peer)
		thread = d.Thread
	default:
		return nil, 0, errors.Errorf("message router must return string or dest: %T", result)
	}

	if err != nil {
		return nil, 0, errors.Wrap(err, "resolve peer")
	}

	return to, thread, nil
}

func (i *iter) resolvePeer(ctx context.Context, peer string) (peers.Peer, error) {
	if peer == "" { // self
		return i.manager.Self(ctx)
	}

	return tutil.GetInputPeer(ctx, i.manager, peer)
}

func (i *iter) resolveCaption(env Env) (*entity.Builder, error) {
	// parse caption
	captionStr, err := texpr.Run(i.caption, env)
	if err != nil {
		return nil, errors.Wrap(err, "parse caption")
	}

	r, ok := captionStr.(string)
	if !ok {
		return nil, errors.Errorf("caption must return string, got %T", captionStr)
	}

	caption := &entity.Builder{}
	if len(r) > 0 {
		if err = html.HTML(strings.NewReader(r), caption, html.Options{
			UserResolver:          nil,
			DisableTelegramEscape: false,
		}); err != nil {
			return nil, errors.Wrap(err, "parse caption HTML")
		}
	}

	return caption, nil
}

func (i *iter) resolveThumb(path string) (*uploaderFile, error) {
	if path == "" {
		return nil, nil
	}

	// has thumbnail
	mime, err := mimetype.DetectFile(path)
	if err != nil || !mediautil.IsImage(mime.String()) { // TODO(iyear): jpg only
		return nil, errors.Wrapf(err, "invalid thumbnail file: %v", path)
	}

	return i.resolveFile(path)
}

func (i *iter) Value() uploader.Elem {
	return i.file
}

func (i *iter) Err() error {
	return i.err
}
