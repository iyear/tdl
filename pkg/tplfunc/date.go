package tplfunc

import (
	"text/template"
	"time"
)

var Date = []Func{Now(), FormatDate()}

func Now() Func {
	return func(funcMap template.FuncMap) {
		funcMap["now"] = func() int64 {
			return time.Now().Unix()
		}
	}
}

func FormatDate() Func {
	return func(funcMap template.FuncMap) {
		funcMap["formatDate"] = func(unix int64) string {
			return time.Unix(unix, 0).Format("20060102150405")
		}
	}
}
