package tplfunc

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"text/template"
)

func stringSlice(args []string) string {
	s := make([]string, len(args))
	for i, v := range args {
		s[i] = fmt.Sprintf(`"%s"`, v)
	}
	return strings.Join(s, " ")
}

func TestRepeat(t *testing.T) {
	tests := []struct {
		name string
		S    string
		N    int
		want string
	}{
		{name: "empty", S: "", N: 0, want: ""},
		{name: "zero", S: "test", N: 0, want: ""},
		{name: "one", S: "test", N: 1, want: "test"},
		{name: "two", S: "test", N: 2, want: "testtest"},
	}

	m := FuncMap(Repeat())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.Builder{}

			err := template.Must(template.New("test").
				Funcs(m).
				Parse(`{{ repeat .S .N }}`)).
				Execute(&got, tt)
			if err != nil {
				t.Errorf("repeat() error = %v", err)
				return
			}
			if got.String() != tt.want {
				t.Errorf("repeat() got = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestReplace(t *testing.T) {
	tests := []struct {
		name  string
		S     string
		Pairs []string
		want  string
	}{
		{name: "empty", S: "", Pairs: nil, want: ""},
		{name: "empty pairs", S: "test", Pairs: nil, want: "test"},
		{name: "single pair", S: "test", Pairs: []string{"t", "T"}, want: "TesT"},
		{name: "multiple pairs1", S: "test", Pairs: []string{"t", "T", "e", "E"}, want: "TEsT"},
		{name: "multiple pairs2", S: "test", Pairs: []string{"t", "T", "e", "E", "s", "S"}, want: "TEST"},
	}

	m := FuncMap(Replace())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.Builder{}

			err := template.Must(template.New("test").
				Funcs(m).
				Parse(fmt.Sprintf(`{{ replace .S %s }}`, stringSlice(tt.Pairs)))).
				Execute(&got, tt)
			if err != nil {
				t.Errorf("replace() error = %v", err)
				return
			}
			if got.String() != tt.want {
				t.Errorf("replace() got = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestReplacePanic(t *testing.T) {
	tests := []struct {
		name  string
		S     string
		Pairs []string
	}{
		{name: "odd pairs", S: "test", Pairs: []string{"t", "T", "e"}},
	}

	m := FuncMap(Replace())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := template.Must(template.New("test").
				Funcs(m).
				Parse(fmt.Sprintf(`{{ replace .S %s }}`, stringSlice(tt.Pairs)))).
				Execute(io.Discard, tt)
			if err == nil {
				t.Errorf("replace() expected error")
			}
		})
	}
}

func TestToUpper(t *testing.T) {
	tests := []struct {
		name string
		S    string
		want string
	}{
		{name: "empty", S: "", want: ""},
		{name: "lower", S: "test", want: "TEST"},
		{name: "upper", S: "TEST", want: "TEST"},
	}

	m := FuncMap(ToUpper())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.Builder{}

			err := template.Must(template.New("test").
				Funcs(m).
				Parse(`{{ upper .S }}`)).
				Execute(&got, tt)
			if err != nil {
				t.Errorf("upper() error = %v", err)
				return
			}
			if got.String() != tt.want {
				t.Errorf("upper() got = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestToLower(t *testing.T) {
	tests := []struct {
		name string
		S    string
		want string
	}{
		{name: "empty", S: "", want: ""},
		{name: "lower", S: "test", want: "test"},
		{name: "upper", S: "TEST", want: "test"},
	}

	m := FuncMap(ToLower())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.Builder{}

			err := template.Must(template.New("test").
				Funcs(m).
				Parse(`{{ lower .S }}`)).
				Execute(&got, tt)
			if err != nil {
				t.Errorf("lower() error = %v", err)
				return
			}
			if got.String() != tt.want {
				t.Errorf("lower() got = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestSnakeCase(t *testing.T) {
	tests := []struct {
		name string
		S    string
		want string
	}{
		{name: "empty", S: "", want: ""},
		{name: "lower", S: "test", want: "test"},
		{name: "upper", S: "TEST", want: "test"},
		{name: "camel", S: "testTest", want: "test_test"},
		{name: "pascal", S: "TestTest", want: "test_test"},
	}

	m := FuncMap(SnakeCase())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.Builder{}

			err := template.Must(template.New("test").
				Funcs(m).
				Parse(`{{ snakecase .S }}`)).
				Execute(&got, tt)
			if err != nil {
				t.Errorf("snakecase() error = %v", err)
				return
			}
			if got.String() != tt.want {
				t.Errorf("snakecase() got = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestCamelCase(t *testing.T) {
	tests := []struct {
		name string
		S    string
		want string
	}{
		{name: "empty", S: "", want: ""},
		{name: "lower", S: "test", want: "Test"},
		{name: "upper", S: "TEST", want: "Test"},
		{name: "snake", S: "test_test", want: "TestTest"},
		{name: "pascal", S: "TestTest", want: "TestTest"},
	}

	m := FuncMap(CamelCase())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.Builder{}

			err := template.Must(template.New("test").
				Funcs(m).
				Parse(`{{ camelcase .S }}`)).
				Execute(&got, tt)
			if err != nil {
				t.Errorf("camelcase() error = %v", err)
				return
			}
			if got.String() != tt.want {
				t.Errorf("camelcase() got = %v, want %v", got.String(), tt.want)
			}
		})
	}
}

func TestKebabCase(t *testing.T) {
	tests := []struct {
		name string
		S    string
		want string
	}{
		{name: "empty", S: "", want: ""},
		{name: "lower", S: "test", want: "test"},
		{name: "upper", S: "TEST", want: "test"},
		{name: "camel", S: "testTest", want: "test-test"},
		{name: "pascal", S: "TestTest", want: "test-test"},
	}

	m := FuncMap(KebabCase())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strings.Builder{}

			err := template.Must(template.New("test").
				Funcs(m).
				Parse(`{{ kebabcase .S }}`)).
				Execute(&got, tt)
			if err != nil {
				t.Errorf("kebabcase() error = %v", err)
				return
			}
			if got.String() != tt.want {
				t.Errorf("kebabcase() got = %v, want %v", got.String(), tt.want)
			}
		})
	}
}
