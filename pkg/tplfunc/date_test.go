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

func TestCustomFormat(t *testing.T) {
	// unify time zone
	time.Local = time.UTC

	tests := []struct {
		name   string
		Unix   int64
		Format string
		want   string
	}{
		{name: "formatDate1", Unix: 0, Format: "2006-01-02 15:04:05", want: "1970-01-01 00:00:00"},
		{name: "formatDate2", Unix: 1, Format: "2006-01-02 15:04:05", want: "1970-01-01 00:00:01"},
		{name: "formatDate3", Unix: 1000000000, Format: "2006-01-02 15:04:05", want: "2001-09-09 01:46:40"},
		{name: "formatDate4", Unix: 0, Format: "Mon, 02 Jan 2006 15:04:05 MST", want: "Thu, 01 Jan 1970 00:00:00 UTC"},
		{name: "formatDate5", Unix: 1, Format: "Mon, 02 Jan 2006 15:04:05 MST", want: "Thu, 01 Jan 1970 00:00:01 UTC"},
		{name: "formatDate6", Unix: 1000000000, Format: "Mon, 02 Jan 2006 15:04:05 MST", want: "Sun, 09 Sep 2001 01:46:40 UTC"},
		{name: "formatDate7", Unix: 0, Format: "20060102150405", want: "19700101000000"},
		{name: "formatDate8", Unix: 1, Format: "20060102150405", want: "19700101000001"},
		{name: "formatDate9", Unix: 1000000000, Format: "20060102150405", want: "20010909014640"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := strings.Builder{}

			err := template.Must(template.New("test").
				Funcs(FuncMap(FormatDate())).
				Parse(fmt.Sprintf(`{{ formatDate %v "%v" }}`, tt.Unix, tt.Format))).
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
