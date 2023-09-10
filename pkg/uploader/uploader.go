package uploader

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/fatih/color"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/telegram/message/styling"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/telegram/uploader"
	"github.com/gotd/td/tg"
	"github.com/jedib0t/go-pretty/v6/progress"
	"golang.org/x/sync/errgroup"

	"github.com/iyear/tdl/pkg/kv"
	"github.com/iyear/tdl/pkg/prog"
	"github.com/iyear/tdl/pkg/storage"
	"github.com/iyear/tdl/pkg/utils"
)

var formatter = utils.Byte.FormatBinaryBytes

type Uploader struct {
	pw   progress.Writer
	opts Options
}

type Options struct {
	Client   *tg.Client
	KV       kv.KV
	PartSize int
	Threads  int
	Iter     Iter
	Photo    bool
}

func New(o Options) *Uploader {
	return &Uploader{
		pw:   prog.New(formatter),
		opts: o,
	}
}

func (u *Uploader) to(ctx context.Context, chat string) (peers.Peer, error) {
	manager := peers.Options{Storage: storage.NewPeers(u.opts.KV)}.Build(u.opts.Client)
	if chat == "" {
		return manager.FromInputPeer(ctx, &tg.InputPeerSelf{})
	}

	return utils.Telegram.GetInputPeer(ctx, manager, chat)
}

func (u *Uploader) Upload(ctx context.Context, chat string, limit int) error {
	to, err := u.to(ctx, chat)
	if err != nil {
		return err
	}

	u.pw.Log(color.GreenString("All files will be uploaded to '%s' dialog", to.VisibleName()))

	u.pw.SetNumTrackersExpected(u.opts.Iter.Total(ctx))

	go u.pw.Render()

	wg, errctx := errgroup.WithContext(ctx)
	wg.SetLimit(limit)

	go runPS(errctx, u.pw)

	for u.opts.Iter.Next(ctx) {
		item, err := u.opts.Iter.Value(ctx)
		if err != nil {
			u.pw.Log(color.RedString("Get item failed: %v, skip...", err))
			continue
		}

		wg.Go(func() error {
			if err := u.upload(errctx, to.InputPeer(), item); err != nil {
				return fmt.Errorf("upload failed: %w", err)
			}

			// remove here so file has been closed in upload function
			u.opts.Iter.Finish(ctx, item.ID)
			return nil
		})
	}

	err = wg.Wait()
	if err != nil {
		u.pw.Stop()
		for u.pw.IsRenderInProgress() {
			time.Sleep(time.Millisecond * 10)
		}

		if errors.Is(err, context.Canceled) {
			color.Red("Upload aborted.")
		}
		return err
	}

	prog.Wait(u.pw)

	return nil
}

func (u *Uploader) upload(ctx context.Context, to tg.InputPeerClass, item *Item) error {
	defer func(r io.ReadCloser, t io.ReadCloser) {
		_ = r.Close()
		_ = t.Close()
	}(item.File, item.Thumb)

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	tracker := prog.AppendTracker(u.pw, formatter, item.Name, item.Size)

	up := uploader.NewUploader(u.opts.Client).
		WithPartSize(u.opts.PartSize).WithThreads(u.opts.Threads).WithProgress(&_progress{tracker: tracker})

	f, err := up.Upload(ctx, uploader.NewUpload(item.Name, item.File, item.Size))
	if err != nil {
		return err
	}

	caption := []message.StyledTextOption{
		styling.Code(item.Name),
		styling.Plain(" - "),
		styling.Code(item.MIME),
	}
	doc := message.UploadedDocument(f, caption...).MIME(item.MIME).Filename(item.Name)
	// upload thumbnail TODO(iyear): maybe still unavailable
	if thumb, err := uploader.NewUploader(u.opts.Client).
		FromReader(ctx, fmt.Sprintf("%s.thumb", item.Name), item.Thumb); err == nil {
		doc = doc.Thumb(thumb)
	}

	var media message.MediaOption = doc
	// upload as photo
	if utils.Media.IsImage(item.MIME) && u.opts.Photo {
		media = message.UploadedPhoto(f, caption...)
	} else if utils.Media.IsVideo(item.MIME) {
		// reset reader
		if _, err = item.File.Seek(0, io.SeekStart); err != nil {
			return err
		}
		dur, w, h, err := utils.Media.GetMP4Info(item.File)
		if err != nil {
			// #132. There may be some errors, but we can still upload the file
			u.pw.Log(color.RedString("Get MP4 information failed: %v, skip set duration and resolution", err))
		} else {
			media = doc.Video().Duration(time.Duration(dur)*time.Second).Resolution(w, h).SupportsStreaming()
		}
	} else if utils.Media.IsAudio(item.MIME) {
		media = doc.Audio().Title(utils.FS.GetNameWithoutExt(item.Name))
	}

	_, err = message.NewSender(u.opts.Client).WithUploader(up).To(to).Media(ctx, media)
	if err != nil {
		return fmt.Errorf("send message failed: %w", err)
	}

	return nil
}
