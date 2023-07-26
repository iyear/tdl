package tplfunc

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"text/template"
	"time"
)

func TestNow(t *testing.T) {
	b := strings.Builder{}
	err := template.Must(template.New("test").
		Funcs(FuncMap(Now())).
		Parse(`{{ now }}`)).
		Execute(&b, nil)
	if err != nil {
		t.Errorf("now() error = %v", err)
		return
	}

	n, err := strconv.ParseInt(b.String(), 10, 64)
	if err != nil {
		t.Errorf("now() error = %v", err)
		return
	}

	// Allow for a second of drift.
	if time.Now().Unix()-n > 1 {
		t.Errorf("now() got = %v", n)
	}
}

func TestFormatDate(t *testing.T) {
	// unify time zone
	time.Local = time.UTC

	tests := []struct {
		name string
		Unix int64
		want string
	}{
		{name: "formatDate1", Unix: 0, want: "19700101000000"},
		{name: "formatDate2", Unix: 1, want: "19700101000001"},
		{name: "formatDate3", Unix: 1000000000, want: "20010909014640"},
		{name: "formatDate4", Unix: 10000000000, want: "22861120174640"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := strings.Builder{}

			err := template.Must(template.New("test").
				Funcs(FuncMap(FormatDate())).
				Parse(fmt.Sprintf(`{{ formatDate %v }}`, tt.Unix))).
				Execute(&b, tt)
			if err != nil {
				t.Errorf("formatDate() error = %v", err)
				return
			}

			if b.String() != tt.want {
				t.Errorf("formatDate() got = %v, want %v", b.String(), tt.want)
			}
		})
	}
}
