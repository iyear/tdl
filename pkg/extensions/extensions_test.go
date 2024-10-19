package extensions

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBaseExtension(t *testing.T) {
	tests := []struct {
		name         string
		path         string
		expectedName string
	}{
		{
			name:         "with prefix",
			path:         "/path/to/tdl-extension",
			expectedName: "extension",
		},
		{
			name:         "without prefix",
			path:         "/path/to/extension2",
			expectedName: "extension2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := baseExtension{path: tt.path}
			assert.Equal(t, tt.expectedName, e.Name())
		})
	}
}
