package up

import (
	"context"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/expr-lang/expr/vm"
	"github.com/go-faster/errors"
	"github.com/go-viper/mapstructure/v2"
	"github.com/gotd/td/telegram/message/entity"
	"github.com/gotd/td/telegram/message/html"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/core/uploader"
	"github.com/iyear/tdl/core/util/tutil"
	"github.com/iyear/tdl/pkg/texpr"
)

type uploadTask struct {
	media      []uploader.MediaDescriptor
	to         peers.Peer
	thread     int
	remove     bool
	paths      []string
	totalSize  int64
	label      string
	entityByID map[string][]tg.MessageEntityClass
}

func (t *uploadTask) captionEntities(media uploader.MediaDescriptor) []tg.MessageEntityClass {
	if t.entityByID == nil {
		return nil
	}
	return t.entityByID[media.Path]
}

type dest struct {
	Peer   string
	Thread int
}

type taskResolver struct {
	to      *vm.Program
	caption *vm.Program
	chat    string
	topic   int
	manager *peers.Manager
}

func (r *taskResolver) resolve(ctx context.Context, file *File) (peers.Peer, int, string, []tg.MessageEntityClass, error) {
	env := exprEnv(ctx, file)

	to, thread, err := r.resolveDest(ctx, env)
	if err != nil {
		return nil, 0, "", nil, errors.Wrap(err, "resolve destination")
	}

	caption, entities, err := r.resolveCaption(env)
	if err != nil {
		return nil, 0, "", nil, errors.Wrap(err, "resolve caption")
	}

	return to, thread, caption, entities, nil
}

func (r *taskResolver) resolveDest(ctx context.Context, env Env) (peers.Peer, int, error) {
	if r.chat != "" {
		to, err := r.resolvePeer(ctx, r.chat)
		if err != nil {
			return nil, 0, errors.Wrap(err, "resolve chat")
		}
		return to, r.topic, nil
	}

	result, err := texpr.Run(r.to, env)
	if err != nil {
		return nil, 0, errors.Wrap(err, "parse expression")
	}

	var (
		to     peers.Peer
		thread int
	)

	switch out := result.(type) {
	case string:
		to, err = r.resolvePeer(ctx, out)
	case map[string]interface{}:
		var d dest
		if err = mapstructure.WeakDecode(out, &d); err != nil {
			return nil, 0, errors.Wrapf(err, "decode dest: %v", result)
		}
		to, err = r.resolvePeer(ctx, d.Peer)
		thread = d.Thread
	default:
		return nil, 0, errors.Errorf("message router must return string or dest: %T", result)
	}

	if err != nil {
		return nil, 0, errors.Wrap(err, "resolve peer")
	}

	return to, thread, nil
}

func (r *taskResolver) resolvePeer(ctx context.Context, peer string) (peers.Peer, error) {
	if peer == "" {
		return r.manager.Self(ctx)
	}
	return tutil.GetInputPeer(ctx, r.manager, peer)
}

func (r *taskResolver) resolveCaption(env Env) (string, []tg.MessageEntityClass, error) {
	captionResult, err := texpr.Run(r.caption, env)
	if err != nil {
		return "", nil, errors.Wrap(err, "parse caption")
	}

	raw, ok := captionResult.(string)
	if !ok {
		return "", nil, errors.Errorf("caption must return string, got %T", captionResult)
	}

	if raw == "" {
		return "", nil, nil
	}

	eb := &entity.Builder{}
	if err = html.HTML(strings.NewReader(raw), eb, html.Options{
		UserResolver:          nil,
		DisableTelegramEscape: false,
	}); err != nil {
		return "", nil, errors.Wrap(err, "parse caption HTML")
	}

	msg, entities := eb.Complete()
	return msg, entities, nil
}

type iter struct {
	tasks []*uploadTask
	delay time.Duration

	cur  int
	err  error
	elem uploader.Elem
}

func newIter(tasks []*uploadTask, delay time.Duration) *iter {
	return &iter{
		tasks: tasks,
		delay: delay,
	}
}

func (i *iter) Next(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		i.err = ctx.Err()
		return false
	default:
	}

	if i.cur >= len(i.tasks) || i.err != nil {
		return false
	}

	if i.delay > 0 && i.cur > 0 {
		time.Sleep(i.delay)
	}

	task := i.tasks[i.cur]
	i.cur++

	i.elem = &iterElem{
		task: task,
	}
	return true
}

func (i *iter) Value() uploader.Elem {
	return i.elem
}

func (i *iter) Err() error {
	return i.err
}

func ext(path string) string {
	return strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
}

func isDetectVideoExt(path string) bool {
	return slices.Contains([]string{"mp4", "m4v", "mov"}, ext(path))
}
