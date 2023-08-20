package chat

import (
	"testing"

	"github.com/iyear/tdl/pkg/texpr"
)

func TestMessageExpr(t *testing.T) {
	msg := &message{
		Mentioned:     true,
		Silent:        false,
		FromScheduled: true,
		Pinned:        false,
		ID:            100,
		FromID:        200,
		Date:          1684651590,
		Message:       "Hello World",
		Media:         messageMedia{Size: 10240, Name: "foo.zip", DC: 3},
		Views:         200,
		Forwards:      100,
	}

	tests := []struct {
		name     string
		expr     string
		expected bool
	}{
		{
			name:     "and",
			expr:     "Mentioned && ID==100 && Date>1684650000",
			expected: true,
		},
		{
			name:     "or",
			expr:     "Mentioned || ID<1000 || Views>100",
			expected: true,
		},
		{
			name:     "match file name .zip extension",
			expr:     `Media.Name matches ".*\\.zip"`,
			expected: true,
		},
		{
			name:     "match file name .zip extension2",
			expr:     `Media.Name endsWith ".zip"`,
			expected: true,
		},
		{
			name:     "match file name and DC",
			expr:     `Media.Name matches "foo*" && Media.DC==3`,
			expected: true,
		},
		{
			name:     "file name contains",
			expr:     `Media.Name contains "foo"`,
			expected: true,
		},
		{
			name:     "match file size",
			expr:     `Media.Size > 5*1024`,
			expected: true,
		},
		{
			name:     "false",
			expr:     `Media.Size > 20*1024 || Media.DC==2 || Silent`,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			expr, err := texpr.Compile(test.expr)
			if err != nil {
				t.Fatal(err)
			}

			got, err := texpr.Run(expr, msg)
			if err != nil {
				t.Fatal(err)
			}

			if got != test.expected {
				t.Errorf("name: %s, expected: %v, got: %v", test.name, test.expected, got)
			}
		})
	}
}
