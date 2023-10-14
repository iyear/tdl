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
		funcMap["formatDate"] = func(args ...any) string {
			switch len(args) {
			case 0:
				panic("formatDate() requires at least 1 argument")
			case 1:
				return time.Unix(int64(args[0].(int)), 0).Format("20060102150405")
			case 2:
				return time.Unix(int64(args[0].(int)), 0).Format(args[1].(string))
			default:
				panic("formatDate() requires at most 2 arguments")
			}
		}
	}
}
