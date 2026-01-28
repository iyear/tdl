package forward

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveRenameFile(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantNil bool
		wantErr bool
	}{
		{
			name:    "empty input returns nil",
			input:   "",
			wantNil: true,
			wantErr: false,
		},
		{
			name:    "valid expression compiles",
			input:   `"test_" + Message.Media.Name`,
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "simple string expression",
			input:   `"renamed_file.mp4"`,
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "complex expression with string conversion",
			input:   "`[` + string(From.ID) + `_` + string(Message.ID) + `]_` + Message.Media.Name",
			wantNil: false,
			wantErr: false,
		},
		{
			name:    "invalid expression fails",
			input:   "this is not valid expr {{{}}}",
			wantNil: true, // when error occurs, result is nil
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := resolveRenameFile(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("resolveRenameFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("resolveRenameFile() got = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}

func TestResolveRenameFileFromFile(t *testing.T) {
	// Create temp file with expression
	tmpDir := t.TempDir()
	exprFile := filepath.Join(tmpDir, "rename_expr.txt")
	expr := "`[` + string(From.ID) + `]_` + Message.Media.Name"

	if err := os.WriteFile(exprFile, []byte(expr), 0o644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	// Test reading from file
	got, err := resolveRenameFile(exprFile)
	if err != nil {
		t.Errorf("resolveRenameFile() from file error = %v", err)
		return
	}
	if got == nil {
		t.Error("resolveRenameFile() from file returned nil, expected compiled program")
	}
}
