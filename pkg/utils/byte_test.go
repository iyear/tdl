package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_byte_ParseBinaryBytes(t *testing.T) {
	tests := []struct {
		name    string
		s       string
		want    int64
		wantErr bool
	}{
		{name: "bytes", s: "100B", want: 100, wantErr: false},
		{name: "bytes lower", s: "100b", want: 100, wantErr: false},
		{name: "kilobytes", s: "1KB", want: 1024, wantErr: false},
		{name: "kilobytes lower", s: "1kb", want: 1024, wantErr: false},
		{name: "megabytes", s: "10MB", want: 10 * 1024 * 1024, wantErr: false},
		{name: "gigabytes", s: "1.5GB", want: int64(1.5 * 1024 * 1024 * 1024), wantErr: false},
		{name: "raw number", s: "100", want: 100, wantErr: false},
		{name: "invalid unit", s: "100ZB", want: 0, wantErr: true},
		{name: "invalid format", s: "abc", want: 0, wantErr: true},
		{name: "empty", s: "", want: 0, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Byte.ParseBinaryBytes(tt.s)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
