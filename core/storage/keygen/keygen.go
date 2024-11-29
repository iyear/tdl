package keygen

import (
	"bytes"
	"strings"
	"sync"
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
