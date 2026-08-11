package mediautil

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

const (
	ebmlSegment        = 0x18538067
	ebmlInfo           = 0x1549a966
	ebmlTimestampScale = 0x2ad7b1
	ebmlDuration       = 0x4489
	ebmlTracks         = 0x1654ae6b
	ebmlTrackEntry     = 0xae
	ebmlTrackType      = 0x83
	ebmlVideo          = 0xe0
	ebmlPixelWidth     = 0xb0
	ebmlPixelHeight    = 0xba

	matroskaVideoTrack = 1
)

type ebmlHeader struct {
	id      uint64
	size    uint64
	unknown bool
}

// GetMatroskaInfo reads duration and dimensions from MKV/WebM container
// metadata. It does not decode video frames or invoke external programs.
func GetMatroskaInfo(r io.ReadSeeker) (duration, width, height int, rerr error) {
	start, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, 0, 0, err
	}
	defer func() {
		if _, err := r.Seek(start, io.SeekStart); rerr == nil && err != nil {
			rerr = err
		}
	}()

	if _, err = r.Seek(0, io.SeekStart); err != nil {
		return 0, 0, 0, err
	}
	fileEnd, err := r.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, 0, 0, err
	}
	if _, err = r.Seek(0, io.SeekStart); err != nil {
		return 0, 0, 0, err
	}

	segmentEnd := fileEnd
	foundSegment := false
	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos >= fileEnd {
			break
		}
		h, err := readEBMLHeader(r)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("read Matroska header: %w", err)
		}
		if h.id == ebmlSegment {
			foundSegment = true
			if !h.unknown {
				segmentEnd, err = elementEnd(r, h.size, fileEnd)
				if err != nil {
					return 0, 0, 0, err
				}
			}
			break
		}
		if h.unknown {
			return 0, 0, 0, fmt.Errorf("unknown-sized element before Matroska segment")
		}
		if err = skipEBML(r, h.size, fileEnd); err != nil {
			return 0, 0, 0, err
		}
	}
	if !foundSegment {
		return 0, 0, 0, fmt.Errorf("Matroska segment not found")
	}

	timestampScale := uint64(1_000_000)
	var rawDuration float64
	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos >= segmentEnd || (rawDuration > 0 && width > 0 && height > 0) {
			break
		}
		h, err := readEBMLHeader(r)
		if err != nil {
			return 0, 0, 0, fmt.Errorf("read Matroska segment: %w", err)
		}
		if h.unknown {
			return 0, 0, 0, fmt.Errorf("unsupported unknown-sized Matroska element 0x%x", h.id)
		}
		end, err := elementEnd(r, h.size, segmentEnd)
		if err != nil {
			return 0, 0, 0, err
		}
		switch h.id {
		case ebmlInfo:
			timestampScale, rawDuration, err = readMatroskaInfo(r, end, timestampScale, rawDuration)
		case ebmlTracks:
			width, height, err = readMatroskaTracks(r, end)
			if err == nil {
				_, err = r.Seek(end, io.SeekStart)
			}
		default:
			_, err = r.Seek(end, io.SeekStart)
		}
		if err != nil {
			return 0, 0, 0, err
		}
	}

	if rawDuration <= 0 || width <= 0 || height <= 0 {
		return 0, 0, 0, fmt.Errorf("incomplete Matroska video metadata")
	}
	seconds := rawDuration * float64(timestampScale) / float64(1_000_000_000)
	return int(seconds), width, height, nil
}

func readMatroskaInfo(r io.ReadSeeker, end int64, scale uint64, duration float64) (uint64, float64, error) {
	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos >= end {
			return scale, duration, nil
		}
		h, err := readEBMLHeader(r)
		if err != nil {
			return 0, 0, err
		}
		switch h.id {
		case ebmlTimestampScale:
			scale, err = readEBMLUint(r, h.size)
		case ebmlDuration:
			duration, err = readEBMLFloat(r, h.size)
		default:
			err = skipEBML(r, h.size, end)
		}
		if err != nil {
			return 0, 0, err
		}
	}
}

func readMatroskaTracks(r io.ReadSeeker, end int64) (int, int, error) {
	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos >= end {
			return 0, 0, nil
		}
		h, err := readEBMLHeader(r)
		if err != nil {
			return 0, 0, err
		}
		entryEnd, err := elementEnd(r, h.size, end)
		if err != nil {
			return 0, 0, err
		}
		if h.id != ebmlTrackEntry {
			if _, err = r.Seek(entryEnd, io.SeekStart); err != nil {
				return 0, 0, err
			}
			continue
		}
		trackType, width, height, err := readMatroskaTrack(r, entryEnd)
		if err != nil {
			return 0, 0, err
		}
		if trackType == matroskaVideoTrack && width > 0 && height > 0 {
			return width, height, nil
		}
	}
}

func readMatroskaTrack(r io.ReadSeeker, end int64) (trackType, width, height int, _ error) {
	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos >= end {
			return trackType, width, height, nil
		}
		h, err := readEBMLHeader(r)
		if err != nil {
			return 0, 0, 0, err
		}
		elemEnd, err := elementEnd(r, h.size, end)
		if err != nil {
			return 0, 0, 0, err
		}
		switch h.id {
		case ebmlTrackType:
			v, err := readEBMLUint(r, h.size)
			if err != nil {
				return 0, 0, 0, err
			}
			trackType = int(v)
		case ebmlVideo:
			width, height, err = readMatroskaVideo(r, elemEnd)
			if err != nil {
				return 0, 0, 0, err
			}
		default:
			if _, err = r.Seek(elemEnd, io.SeekStart); err != nil {
				return 0, 0, 0, err
			}
		}
	}
}

func readMatroskaVideo(r io.ReadSeeker, end int64) (width, height int, _ error) {
	for {
		pos, _ := r.Seek(0, io.SeekCurrent)
		if pos >= end {
			return width, height, nil
		}
		h, err := readEBMLHeader(r)
		if err != nil {
			return 0, 0, err
		}
		switch h.id {
		case ebmlPixelWidth:
			v, err := readEBMLUint(r, h.size)
			if err != nil {
				return 0, 0, err
			}
			width = int(v)
		case ebmlPixelHeight:
			v, err := readEBMLUint(r, h.size)
			if err != nil {
				return 0, 0, err
			}
			height = int(v)
		default:
			if err = skipEBML(r, h.size, end); err != nil {
				return 0, 0, err
			}
		}
	}
}

func readEBMLHeader(r io.Reader) (ebmlHeader, error) {
	id, _, _, err := readVINT(r, false)
	if err != nil {
		return ebmlHeader{}, err
	}
	size, n, unknown, err := readVINT(r, true)
	if err != nil {
		return ebmlHeader{}, err
	}
	if n > 8 {
		return ebmlHeader{}, fmt.Errorf("invalid EBML size length %d", n)
	}
	return ebmlHeader{id: id, size: size, unknown: unknown}, nil
}

func readVINT(r io.Reader, clearMarker bool) (value uint64, length int, unknown bool, err error) {
	var first [1]byte
	if _, err = io.ReadFull(r, first[:]); err != nil {
		return 0, 0, false, err
	}
	mask := byte(0x80)
	for length = 1; length <= 8 && first[0]&mask == 0; length++ {
		mask >>= 1
	}
	if length > 8 {
		return 0, 0, false, fmt.Errorf("invalid EBML variable integer")
	}
	value = uint64(first[0])
	if clearMarker {
		value = uint64(first[0] &^ mask)
	}
	for index := 1; index < length; index++ {
		if _, err = io.ReadFull(r, first[:]); err != nil {
			return 0, 0, false, err
		}
		value = value<<8 | uint64(first[0])
	}
	if clearMarker {
		unknown = value == (uint64(1)<<(7*length))-1
	}
	return value, length, unknown, nil
}

func readEBMLUint(r io.Reader, size uint64) (uint64, error) {
	if size == 0 || size > 8 {
		return 0, fmt.Errorf("invalid EBML integer size %d", size)
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, err
	}
	var value uint64
	for _, b := range buf {
		value = value<<8 | uint64(b)
	}
	return value, nil
}

func readEBMLFloat(r io.Reader, size uint64) (float64, error) {
	switch size {
	case 4:
		var buf [4]byte
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			return 0, err
		}
		return float64(math.Float32frombits(binary.BigEndian.Uint32(buf[:]))), nil
	case 8:
		var buf [8]byte
		if _, err := io.ReadFull(r, buf[:]); err != nil {
			return 0, err
		}
		return math.Float64frombits(binary.BigEndian.Uint64(buf[:])), nil
	default:
		return 0, fmt.Errorf("invalid EBML float size %d", size)
	}
}

func elementEnd(r io.Seeker, size uint64, limit int64) (int64, error) {
	pos, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	if pos > limit || size > uint64(limit-pos) {
		return 0, fmt.Errorf("EBML element exceeds its parent")
	}
	return pos + int64(size), nil
}

func skipEBML(r io.Seeker, size uint64, limit int64) error {
	end, err := elementEnd(r, size, limit)
	if err != nil {
		return err
	}
	_, err = r.Seek(end, io.SeekStart)
	return err
}
