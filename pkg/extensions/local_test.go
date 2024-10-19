package extensions

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLocalExtension(t *testing.T) {
	tests := []struct {
		name        string
		ext         *localExtension
		expectedURL string
	}{
		{
			name: "local 1",
			ext: &localExtension{
				baseExtension: baseExtension{path: "/path/to/local"},
			},
			expectedURL: "file:///path/to/local",
		},
		{
			name: "local 2",
			ext: &localExtension{
				baseExtension: baseExtension{path: "/path/to/local2"},
			},
			expectedURL: "file:///path/to/local2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedURL, tt.ext.URL())
			assert.Equal(t, "local", tt.ext.Owner())
			assert.Equal(t, "", tt.ext.CurrentVersion())
			assert.Equal(t, "", tt.ext.LatestVersion(context.TODO()))
			assert.False(t, tt.ext.UpdateAvailable(context.TODO()))
		})
	}
}
