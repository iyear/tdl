package texpr

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldsGetter(t *testing.T) {
	type T struct {
		F1 int    `comment:"f1 comment"`
		F2 int8   `comment:"f2 comment"`
		F3 string `comment:"f3 comment"`
		F4 struct {
			F5 uint8  `comment:"f5 comment"`
			F6 uint16 `comment:"f6 comment"`
			F7 *struct {
				F8  string            `comment:"f8 comment"`
				F9  map[string]string `comment:"f9 comment"`
				F10 [3]bool           `comment:"f10 comment"`
				F11 []float32         `comment:"f11 comment"`
				F12 []struct {
					F13 int `comment:"f13 comment"`
				}
			}
		}
	}

	fg := NewFieldsGetter(nil)

	fields, err := fg.Walk(&T{})
	require.NoError(t, err)

	expected := `F1: int # f1 comment
F2: int8 # f2 comment
F3: string # f3 comment
F4.F5: uint8 # f5 comment
F4.F6: uint16 # f6 comment
F4.F7.F8: string # f8 comment
F4.F7.F9: map[string]string # f9 comment
F4.F7.F10[]: bool # f10 comment
F4.F7.F11[]: float32 # f11 comment
F4.F7.F12[].F13: int # f13 comment
`
	assert.Equal(t, expected, fg.Sprint(fields, false))
}
