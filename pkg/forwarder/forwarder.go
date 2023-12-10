package forwarder

import (
	"context"
	"math/rand"
	"time"

	"github.com/go-faster/errors"
	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	"go.uber.org/atomic"
	"go.uber.org/zap"

	"github.com/iyear/tdl/pkg/dcpool"
	"github.com/iyear/tdl/pkg/logger"
	"github.com/iyear/tdl/pkg/tmedia"
	"github.com/iyear/tdl/pkg/utils"
)

//go:generate go-enum --values --names --flag --nocase

// Mode
// ENUM(direct, clone)
type Mode int

type Options struct {
	Pool     dcpool.Pool
	PartSize int
	Threads  int
	Iter     Iter
	Progress Progress
}

type Forwarder struct {
	sent map[tuple]struct{} // used to filter grouped messages which are already sent
	rand *rand.Rand
	opts Options
}

type tuple struct {
	from int64
	msg  int
}

func New(opts Options) *Forwarder {
	return &Forwarder{
		sent: make(map[tuple]struct{}),
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
		opts: opts,
	}
}

func (f *Forwarder) Forward(ctx context.Context) error {
	for f.opts.Iter.Next(ctx) {
		elem := f.opts.Iter.Value()
		if _, ok := f.sent[f.tuple(elem.From(), elem.Msg())]; ok {
			// skip grouped messages
			continue
		}

		if _, ok := elem.Msg().GetGroupedID(); ok {
			grouped, err := utils.Telegram.GetGroupedMessages(ctx, f.opts.Pool.Default(ctx), elem.From().InputPeer(), elem.Msg())
			if err != nil {
				continue
			}

			if err = f.forwardMessage(ctx, elem, grouped...); err != nil {
				continue
			}

			continue
		}

		if err := f.forwardMessage(ctx, elem); err != nil {
			// canceled by user, so we directly return error to stop all
			if errors.Is(err, context.Canceled) {
				return err
			}
			continue
		}
	}

	return f.opts.Iter.Err()
}

func (f *Forwarder) forwardMessage(ctx context.Context, elem Elem, grouped ...*tg.Message) (rerr error) {
	f.opts.Progress.OnAdd(elem)
	defer func() {
		f.sent[f.tuple(elem.From(), elem.Msg())] = struct{}{}

		// grouped message also should be marked as sent
		for _, m := range grouped {
			f.sent[f.tuple(elem.From(), m)] = struct{}{}
		}
		f.opts.Progress.OnDone(elem, rerr)
	}()

	log := logger.From(ctx).With(
		zap.Int64("from", elem.From().ID()),
		zap.Int64("to", elem.To().ID()),
		zap.Int("message", elem.Msg().ID))

	// used for clone progress
	totalSize, err := mediaSizeSum(elem.Msg(), grouped...)
	if err != nil {
		return errors.Wrap(err, "media total size")
	}
	done := atomic.NewInt64(0)

	forwardTextOnly := func(msg *tg.Message) error {
		if msg.Message == "" {
			return errors.Errorf("empty message content, skip send: %d", msg.ID)
		}
		req := &tg.MessagesSendMessageRequest{
			NoWebpage:              false,
			Silent:                 elem.AsSilent(),
			Background:             false,
			ClearDraft:             false,
			Noforwards:             false,
			UpdateStickersetsOrder: false,
			Peer:                   elem.To().InputPeer(),
			ReplyTo:                nil,
			Message:                msg.Message,
			RandomID:               f.rand.Int63(),
			ReplyMarkup:            msg.ReplyMarkup,
			Entities:               msg.Entities,
			ScheduleDate:           0,
			SendAs:                 nil,
		}
		req.SetFlags()

		if _, err := f.forwardClient(ctx, elem).MessagesSendMessage(ctx, req); err != nil {
			return errors.Wrap(err, "send message")
		}
		return nil
	}

	convForwardedMedia := func(msg *tg.Message) (tg.InputMediaClass, error) {
		if _, hasMedia := msg.GetMedia(); !hasMedia {
			// media can't be forwarded via simple copy(it depends on the server ids)
			// if it's not a media message, just break and send text copy
			return nil, errors.Errorf("message %d is not a media message", msg.ID)
		}

		// if it's a media message, but it's not protected, convert it to InputMediaClass
		// or if it's protected, but it doesn't contain photo or document,

		// we should clone photo and document via re-upload, it will be banned if we forward it directly.
		// but other media can be forwarded directly via copy
		if (!protectedDialog(elem.From()) && !protectedMessage(msg)) || !photoOrDocument(msg.Media) {
			media, ok := tmedia.ConvInputMedia(msg.Media)
			if !ok {
				return nil, errors.Errorf("can't convert message %d to input class directly", msg.ID)
			}
			return media, nil
		}

		media, ok := tmedia.GetMedia(msg)
		if !ok {
			log.Warn("Can't get media from message",
				zap.Int64("peer", elem.From().ID()),
				zap.Int("message", msg.ID))

			// unsupported re-upload media
			return nil, errors.Errorf("unsupported media %T", msg.Media)
		}

		mediaFile, err := f.cloneMedia(ctx, cloneOptions{
			elem:  elem,
			media: media,
			progress: &wrapProgress{
				elem:     elem,
				progress: f.opts.Progress,
				done:     done,
				total:    totalSize * 2,
			},
		}, elem.AsDryRun())
		if err != nil {
			return nil, errors.Wrap(err, "clone media")
		}

		var inputMedia tg.InputMediaClass
		// now we only have to process cloned photo or document
		switch m := msg.Media.(type) {
		case *tg.MessageMediaPhoto:
			photo := &tg.InputMediaUploadedPhoto{
				Spoiler:    m.Spoiler,
				File:       mediaFile,
				TTLSeconds: m.TTLSeconds,
			}
			photo.SetFlags()

			inputMedia = photo
		case *tg.MessageMediaDocument:
			doc, ok := m.Document.AsNotEmpty()
			if !ok {
				return nil, errors.Errorf("empty document %d", msg.ID)
			}

			document := &tg.InputMediaUploadedDocument{
				NosoundVideo: false, // do not set
				ForceFile:    false, // do not set
				Spoiler:      m.Spoiler,
				File:         mediaFile,
				MimeType:     doc.MimeType,
				Attributes:   doc.Attributes,
				Stickers:     nil, // do not set
				TTLSeconds:   0,   // do not set
			}

			if thumb, ok := tmedia.GetDocumentThumb(doc); ok {
				thumbFile, err := f.cloneMedia(ctx, cloneOptions{
					elem:     elem,
					media:    thumb,
					progress: nopProgress{},
				}, elem.AsDryRun())
				if err != nil {
					return nil, errors.Wrap(err, "clone thumb")
				}

				document.Thumb = thumbFile
			}

			document.SetFlags()

			inputMedia = document
		default:
			return nil, errors.Errorf("unsupported media %T", msg.Media)
		}

		// note that they must be separately uploaded using messages uploadMedia first,
		// using raw inputMediaUploaded* constructors is not supported.
		messageMedia, err := f.forwardClient(ctx, elem).MessagesUploadMedia(ctx, &tg.MessagesUploadMediaRequest{
			Peer:  elem.To().InputPeer(),
			Media: inputMedia,
		})
		if err != nil {
			return nil, errors.Wrap(err, "upload media")
		}

		inputMedia, ok = tmedia.ConvInputMedia(messageMedia)
		if !ok && !elem.AsDryRun() {
			return nil, errors.Errorf("can't convert uploaded media to input class")
		}

		return inputMedia, nil
	}

	switch elem.Mode() {
	case ModeDirect:
		// it can be forwarded via API
		if !protectedDialog(elem.From()) && !protectedMessage(elem.Msg()) {
			builder := message.NewSender(f.forwardClient(ctx, elem)).
				To(elem.To().InputPeer()).CloneBuilder()
			if elem.AsSilent() {
				builder = builder.Silent()
			}

			if len(grouped) > 0 {
				ids := make([]int, 0, len(grouped))
				for _, m := range grouped {
					ids = append(ids, m.ID)
				}

				if _, err := builder.ForwardIDs(elem.From().InputPeer(), ids[0], ids[1:]...).Send(ctx); err != nil {
					goto fallback
				}

				return nil
			}

			if _, err := builder.ForwardIDs(elem.From().InputPeer(), elem.Msg().ID).Send(ctx); err != nil {
				goto fallback
			}
			return nil
		}
	fallback:
		fallthrough
	case ModeClone:
		if len(grouped) > 0 {
			media := make([]tg.InputSingleMedia, 0, len(grouped))
			for _, gm := range grouped {
				m, err := convForwardedMedia(gm)
				if err != nil {
					log.Debug("Can't convert forwarded media", zap.Error(err))
					continue
				}

				single := tg.InputSingleMedia{
					Media:    m,
					RandomID: f.rand.Int63(),
					Message:  gm.Message,
					Entities: gm.Entities,
				}
				single.SetFlags()

				media = append(media, single)
			}

			if len(media) > 0 {
				req := &tg.MessagesSendMultiMediaRequest{
					Silent:                 elem.AsSilent(),
					Background:             false,
					ClearDraft:             false,
					Noforwards:             false,
					UpdateStickersetsOrder: false,
					Peer:                   elem.To().InputPeer(),
					ReplyTo:                nil,
					MultiMedia:             media,
					ScheduleDate:           0,
					SendAs:                 nil,
				}
				req.SetFlags()
				if _, err := f.forwardClient(ctx, elem).MessagesSendMultiMedia(ctx, req); err != nil {
					return errors.Wrap(err, "send multi media")
				}
				return nil
			}

			return forwardTextOnly(elem.Msg())
		}

		media, err := convForwardedMedia(elem.Msg())
		if err != nil {
			log.Debug("Can't convert forwarded media", zap.Error(err))
			return forwardTextOnly(elem.Msg())
		}
		// send text copy with forwarded media
		req := &tg.MessagesSendMediaRequest{
			Silent:                 elem.AsSilent(),
			Background:             false,
			ClearDraft:             false,
			Noforwards:             false,
			UpdateStickersetsOrder: false,
			Peer:                   elem.To().InputPeer(),
			ReplyTo:                nil,
			Media:                  media,
			Message:                elem.Msg().Message,
			RandomID:               rand.Int63(),
			ReplyMarkup:            elem.Msg().ReplyMarkup,
			Entities:               elem.Msg().Entities,
			ScheduleDate:           0,
			SendAs:                 nil,
		}
		req.SetFlags()

		if _, err := f.forwardClient(ctx, elem).MessagesSendMedia(ctx, req); err != nil {
			return errors.Wrap(err, "send single media")
		}
		return nil
	}

	return errors.Errorf("unsupported mode %v", elem.Mode())
}

func (f *Forwarder) tuple(peer peers.Peer, msg *tg.Message) tuple {
	return tuple{
		from: peer.ID(),
		msg:  msg.ID,
	}
}

type nopInvoker struct{}

func (n nopInvoker) Invoke(_ context.Context, _ bin.Encoder, _ bin.Decoder) error {
	return nil
}

type nopProgress struct{}

func (nopProgress) add(_ int64) {}

type wrapProgress struct {
	elem     Elem
	progress ProgressClone
	done     *atomic.Int64
	total    int64
}

func (w *wrapProgress) add(n int64) {
	w.progress.OnClone(w.elem, ProgressState{
		Done:  w.done.Add(n),
		Total: w.total,
	})
}

func (f *Forwarder) forwardClient(ctx context.Context, elem Elem) *tg.Client {
	if elem.AsDryRun() {
		return tg.NewClient(nopInvoker{})
	}

	return f.opts.Pool.Default(ctx)
}

func protectedDialog(peer peers.Peer) bool {
	switch p := peer.(type) {
	case peers.Chat:
		return p.Raw().GetNoforwards()
	case peers.Channel:
		return p.Raw().GetNoforwards()
	}

	return false
}

func protectedMessage(msg *tg.Message) bool {
	return msg.GetNoforwards()
}

func photoOrDocument(media tg.MessageMediaClass) bool {
	switch media.(type) {
	case *tg.MessageMediaPhoto, *tg.MessageMediaDocument:
		return true
	default:
		return false
	}
}

func mediaSizeSum(msg *tg.Message, grouped ...*tg.Message) (int64, error) {
	if len(grouped) > 0 {
		total := int64(0)
		for _, gm := range grouped {
			m, ok := tmedia.GetMedia(gm)
			if !ok {
				return 0, errors.Errorf("can't get media from message %d", gm.ID)
			}
			total += m.Size
		}

		return total, nil
	}

	m, ok := tmedia.GetMedia(msg)
	if !ok { // maybe it's a text only message
		return 0, nil
	}

	return m.Size, nil
}
