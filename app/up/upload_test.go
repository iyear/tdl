package up

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"github.com/gotd/td/bin"
	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"
	pw "github.com/jedib0t/go-pretty/v6/progress"
	"github.com/stretchr/testify/require"
	"github.com/yapingcat/gomedia/go-mp4"

	"github.com/iyear/tdl/core/uploader"
)

func TestGroupedUploadSingleFile(t *testing.T) {
	for _, count := range []int{1, uploader.MaxGroupSize + 1} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			names := make([]string, count)
			for i := range names {
				names[i] = fmt.Sprintf("photo-%d.jpg", i)
			}
			files := uploadFiles(t, names...)
			rpc := &uploadRPC{}

			require.NoError(t, runUpload(t, files, Options{Group: true, Photo: true}, rpc, pw.NewWriter()))
			require.Len(t, rpc.singles, 1)
			require.Equal(t, fmt.Sprintf("photo-%d", count-1), rpc.singles[0].Message)
			require.Len(t, rpc.albums, count/uploader.MaxGroupSize)
			if len(rpc.albums) > 0 {
				require.Len(t, rpc.albums[0].MultiMedia, uploader.MaxGroupSize)
			}
		})
	}
}

func TestGroupedUploadRoutesLastFile(t *testing.T) {
	files := uploadFiles(t, "a.jpg", "b.jpg", "c.jpg")
	opts := Options{Group: true, Photo: true, To: `FileName == "c" ? "target" : ""`}
	rpc := &uploadRPC{}

	require.NoError(t, runUpload(t, files, opts, rpc, pw.NewWriter()))
	require.Len(t, rpc.albums, 1)
	require.IsType(t, &tg.InputPeerSelf{}, rpc.albums[0].Peer)
	require.Len(t, rpc.albums[0].MultiMedia, 2)
	require.Equal(t, "a", rpc.albums[0].MultiMedia[0].Message)
	require.Empty(t, rpc.albums[0].MultiMedia[1].Message)
	require.Len(t, rpc.singles, 1)
	require.Equal(t, "c", rpc.singles[0].Message)
	require.IsType(t, &tg.InputPeerChannel{}, rpc.singles[0].Peer)
	require.Equal(t, int64(1), rpc.singles[0].Peer.(*tg.InputPeerChannel).ChannelID)
}

func TestGroupedUploadMediaTypes(t *testing.T) {
	t.Run("document and video", func(t *testing.T) {
		files := uploadFiles(t, "cover.jpg", "clip.mp4")
		rpc := &uploadRPC{}

		require.NoError(t, runUpload(t, files, Options{Group: true}, rpc, pw.NewWriter()))
		require.Empty(t, rpc.albums)
		require.Len(t, rpc.singles, 2)
		require.Equal(t, "cover", rpc.singles[0].Message)
		require.IsType(t, &tg.InputMediaUploadedDocument{}, rpc.singles[0].Media)
		require.Equal(t, "clip", rpc.singles[1].Message)
		require.True(t, hasVideoAttribute(rpc.singles[1].Media.(*tg.InputMediaUploadedDocument)))
	})

	t.Run("photo and video", func(t *testing.T) {
		media := uploadFiles(t, "cover.jpg", "clip4.mp4", "clip4.thumb")
		files, err := walk([]string{media[0].File, media[1].File}, nil, nil)
		require.NoError(t, err)
		rpc := &uploadRPC{}

		require.NoError(t, runUpload(t, files, Options{Group: true, Photo: true}, rpc, pw.NewWriter()))
		require.Empty(t, rpc.singles)
		require.Len(t, rpc.albums, 1)
		require.Len(t, rpc.albums[0].MultiMedia, 2)
		require.Len(t, rpc.uploaded, 2)
		require.IsType(t, &tg.InputMediaUploadedPhoto{}, rpc.uploaded[0])
		require.IsType(t, &tg.InputMediaUploadedDocument{}, rpc.uploaded[1])
		video := rpc.uploaded[1].(*tg.InputMediaUploadedDocument)
		require.True(t, hasVideoAttribute(video))
		require.True(t, video.NosoundVideo)
		require.IsType(t, &tg.InputFile{}, video.Thumb)
		require.Equal(t, "clip4.thumb", video.Thumb.(*tg.InputFile).Name)
	})
}

func TestGroupedUploadRemoveAfterSend(t *testing.T) {
	rejected := errors.New("MEDIA_INVALID")
	for _, tt := range []struct {
		name    string
		album   bool
		sendErr error
	}{
		{name: "album accepted", album: true},
		{name: "album rejected", album: true, sendErr: rejected},
		{name: "single rejected", sendErr: rejected},
	} {
		t.Run(tt.name, func(t *testing.T) {
			names := []string{"a.jpg"}
			if tt.album {
				names = append(names, "b.jpg")
			}
			files := uploadFiles(t, names...)
			writer := &uploadWriter{Writer: pw.NewWriter()}
			rpc := &uploadRPC{sendErr: tt.sendErr}
			rpc.beforeSend = func() error {
				for _, file := range files {
					if _, err := os.Stat(file.File); err != nil {
						return err
					}
				}
				for _, tracker := range writer.trackers {
					if tracker.IsDone() {
						return errors.New("progress completed before send")
					}
				}
				return nil
			}

			require.NoError(t, runUpload(t, files, Options{Group: true, Remove: true}, rpc, writer))
			require.NoError(t, rpc.beforeSendErr)

			require.Len(t, writer.trackers, len(files))
			for i, file := range files {
				require.True(t, writer.trackers[i].IsDone())
				require.Equal(t, tt.sendErr != nil, writer.trackers[i].IsErrored())
				if tt.sendErr == nil {
					require.NoFileExists(t, file.File)
				} else {
					require.FileExists(t, file.File)
				}
			}
			if tt.album {
				require.Len(t, rpc.albums, 1)
				require.Len(t, rpc.uploaded, 2)
				for _, media := range rpc.uploaded {
					require.IsType(t, &tg.InputMediaUploadedDocument{}, media)
					require.True(t, media.(*tg.InputMediaUploadedDocument).ForceFile)
				}
			} else {
				require.Len(t, rpc.singles, 1)
			}
		})
	}
}

func TestUploadContinuesAfterSendFailure(t *testing.T) {
	files := uploadFiles(t, "a.jpg", "b.jpg")
	rpc := &uploadRPC{}
	rpc.beforeSend = func() error {
		if len(rpc.singles) == 1 {
			return errors.New("MEDIA_INVALID")
		}
		return nil
	}

	require.NoError(t, runUpload(t, files, Options{}, rpc, pw.NewWriter()))
	require.Len(t, rpc.singles, 2)
	require.Equal(t, "b", rpc.singles[1].Message)
}

func uploadFiles(t *testing.T, names ...string) []*File {
	t.Helper()

	dir := t.TempDir()
	var photo bytes.Buffer
	require.NoError(t, jpeg.Encode(&photo, image.NewGray(image.Rect(0, 0, 32, 32)), nil))

	files := make([]*File, 0, len(names))
	for _, name := range names {
		path := filepath.Join(dir, name)
		if filepath.Ext(name) == ".mp4" {
			writeTestVideo(t, path)
		} else {
			require.NoError(t, os.WriteFile(path, photo.Bytes(), 0o600))
		}
		files = append(files, &File{File: path})
	}
	return files
}

func writeTestVideo(t *testing.T, path string) {
	t.Helper()

	// Annex B SPS, PPS, and one 32x32 black H.264 keyframe.
	frame := []byte{
		0, 0, 0, 1, 0x67, 0x42, 0xc0, 0x0a, 0xda, 0x25, 0xb0, 0x11, 0, 0, 3,
		0, 1, 0, 0, 3, 0, 2, 0x0f, 0x12, 0x26, 0xa0,
		0, 0, 0, 1, 0x68, 0xce, 0x0f, 0xc8,
		0, 0, 1, 0x65, 0x88, 0x84, 0x3a, 0x26, 0x28, 0, 9, 2, 0xc9, 0xd7, 0x5e,
	}
	file, err := os.Create(path)
	require.NoError(t, err)
	defer func() { require.NoError(t, file.Close()) }()

	muxer, err := mp4.CreateMp4Muxer(file)
	require.NoError(t, err)
	track := muxer.AddVideoTrack(mp4.MP4_CODEC_H264)
	require.NoError(t, muxer.Write(track, frame, 0, 0))
	require.NoError(t, muxer.Write(track, frame, 1000, 1000))
	require.NoError(t, muxer.WriteTrailer())
}

func runUpload(t *testing.T, files []*File, opts Options, rpc *uploadRPC, writer pw.Writer) error {
	t.Helper()

	ctx := t.Context()
	client := tg.NewClient(rpc)
	manager := (peers.Options{}).Build(client)
	to, err := resolveDest(ctx, manager, opts.To)
	require.NoError(t, err)
	caption, err := resolveCaption("FileName")
	require.NoError(t, err)

	up := uploader.New(uploader.Options{
		Client:   client,
		Threads:  1,
		Iter:     newIter(files, to, caption, "", 0, opts.Photo, opts.Remove, opts.Group, 0, manager),
		Progress: newProgress(writer),
	})
	return up.Upload(ctx, 1)
}

func hasVideoAttribute(doc *tg.InputMediaUploadedDocument) bool {
	for _, attr := range doc.Attributes {
		if _, ok := attr.(*tg.DocumentAttributeVideo); ok {
			return true
		}
	}
	return false
}

type uploadWriter struct {
	pw.Writer
	trackers []*pw.Tracker
}

func (w *uploadWriter) AppendTracker(tracker *pw.Tracker) {
	w.trackers = append(w.trackers, tracker)
	w.Writer.AppendTracker(tracker)
}

type uploadRPC struct {
	singles       []*tg.MessagesSendMediaRequest
	albums        []*tg.MessagesSendMultiMediaRequest
	uploaded      []tg.InputMediaClass
	beforeSend    func() error
	beforeSendErr error
	sendErr       error
}

func (r *uploadRPC) Invoke(_ context.Context, input bin.Encoder, output bin.Decoder) error {
	switch req := input.(type) {
	case *tg.UsersGetUsersRequest:
		output.(*tg.UserClassVector).Elems = []tg.UserClass{&tg.User{ID: 1, Self: true, FirstName: "Self"}}
	case *tg.ContactsResolveUsernameRequest:
		if req.Username != "target" {
			return fmt.Errorf("unexpected username %q", req.Username)
		}
		// A channel and a user can have the same numeric ID.
		*output.(*tg.ContactsResolvedPeer) = tg.ContactsResolvedPeer{
			Peer:  &tg.PeerChannel{ChannelID: 1},
			Chats: []tg.ChatClass{&tg.Channel{ID: 1, Title: "Target"}},
		}
	case *tg.UploadSaveFilePartRequest:
		output.(*tg.BoolBox).Bool = &tg.BoolTrue{}
	case *tg.MessagesUploadMediaRequest:
		r.uploaded = append(r.uploaded, req.Media)
		id := int64(len(r.uploaded))
		switch req.Media.(type) {
		case *tg.InputMediaUploadedPhoto:
			output.(*tg.MessageMediaBox).MessageMedia = &tg.MessageMediaPhoto{Photo: &tg.Photo{ID: id}}
		case *tg.InputMediaUploadedDocument:
			output.(*tg.MessageMediaBox).MessageMedia = &tg.MessageMediaDocument{Document: &tg.Document{ID: id}}
		default:
			return fmt.Errorf("unexpected media: %T", req.Media)
		}
	case *tg.MessagesSendMediaRequest:
		r.singles = append(r.singles, req)
		return r.send(output)
	case *tg.MessagesSendMultiMediaRequest:
		r.albums = append(r.albums, req)
		return r.send(output)
	default:
		return fmt.Errorf("unexpected RPC: %T", input)
	}
	return nil
}

func (r *uploadRPC) send(output bin.Decoder) error {
	if r.beforeSend != nil {
		r.beforeSendErr = r.beforeSend()
		if r.beforeSendErr != nil {
			return r.beforeSendErr
		}
	}
	output.(*tg.UpdatesBox).Updates = &tg.Updates{}
	return r.sendErr
}
