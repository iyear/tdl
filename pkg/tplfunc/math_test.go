package tplfunc

import (
	"strconv"
	"strings"
	"testing"
	"text/template"
)

func TestRand(t *testing.T) {
	tests := []struct {
		name string
		Min  int
		Max  int
	}{
		{name: "rand1", Min: 0, Max: 100},
		{name: "rand2", Min: 99, Max: 100},
		{name: "rand3", Min: 95, Max: 100},
		{name: "rand5", Min: 0, Max: 1},
	}

	m := FuncMap(Rand())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.Builder{}

			err := template.Must(template.New("test").
				Funcs(m).
				Parse(`{{ rand .Min .Max }}`)).
				Execute(&got, tt)
			if err != nil {
				t.Errorf("rand() error = %v", err)
				return
			}

			n, err := strconv.Atoi(got.String())
			if err != nil {
				t.Errorf("rand() error = %v", err)
				return
			}

			if n < tt.Min || n > tt.Max {
				t.Errorf("rand() got = %v", n)
			}
		})
	}
}

func TestRandPanic(t *testing.T) {
	m := FuncMap(Rand())
	got := strings.Builder{}

	err := template.Must(template.New("test").
		Funcs(m).
		Parse(`{{ rand .Min .Max }}`)).
		Execute(&got, struct {
			Min int
			Max int
		}{Min: 0, Max: 0})
	if err == nil {
		t.Errorf("rand() expected error, got nil")
		return
	}
}
