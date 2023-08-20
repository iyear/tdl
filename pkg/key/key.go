package key

import (
	"bytes"
	"strconv"
	"strings"
	"sync"

	"github.com/gotd/td/telegram/peers"
)

var keyPool = sync.Pool{
	New: func() interface{} {
		b := &bytes.Buffer{}
		b.Grow(16)
		return b
	},
}

func New(indexes ...string) string {
	buf := keyPool.Get().(*bytes.Buffer)
	buf.WriteString(strings.Join(indexes, ":"))

	t := buf.String()
	buf.Reset()
	keyPool.Put(buf)
	return t
}

func Session() string {
	return New("session")
}

func App() string {
	return New("app")
}

func State(userID int64) string {
	return New("state", strconv.FormatInt(userID, 10))
}

func StateChannel(userID int64) string {
	return New("chan", strconv.FormatInt(userID, 10))
}

func PeersKey(key peers.Key) string {
	return New("peers", "key", key.Prefix, strconv.FormatInt(key.ID, 10))
}

func PeersPhone(phone string) string {
	return New("peers", "phone", phone)
}

func PeersContactsHash() string {
	return New("peers", "contacts", "hash")
}

func Resume(fingerprint string) string {
	return New("resume", fingerprint)
}
