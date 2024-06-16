package dl

import (
	"io"
	"os"

	"github.com/gotd/td/telegram/peers"
	"github.com/gotd/td/tg"

	"github.com/iyear/tdl/core/downloader"
	"github.com/iyear/tdl/core/tmedia"
)

type iterElem struct {
	id int

	from    peers.Peer
	fromMsg *tg.Message
	file    *tmedia.Media

	to *os.File

	opts Options
}

func (i *iterElem) File() downloader.File { return i }

func (i *iterElem) To() io.WriterAt { return i.to }

func (i *iterElem) AsTakeout() bool { return i.opts.Takeout }

func (i *iterElem) Location() tg.InputFileLocationClass { return i.file.InputFileLoc }

func (i *iterElem) Name() string { return i.file.Name }

func (i *iterElem) Size() int64 { return i.file.Size }

func (i *iterElem) DC() int { return i.file.DC }
