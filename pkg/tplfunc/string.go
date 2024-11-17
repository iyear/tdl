package tplfunc

import (
	"strings"
	"text/template"

	"github.com/flytam/filenamify"
	"github.com/iancoleman/strcase"
)

var String = []Func{
	Repeat(), Replace(),
	ToUpper(), ToLower(),
	SnakeCase(), CamelCase(), KebabCase(),
	Filenamify(),
}

func Repeat() Func {
	return func(funcMap template.FuncMap) {
		funcMap["repeat"] = func(s string, n int) string {
			return strings.Repeat(s, n)
		}
	}
}

func Replace() Func {
	return func(funcMap template.FuncMap) {
		funcMap["replace"] = func(s string, pairs ...string) string {
			return strings.NewReplacer(pairs...).Replace(s)
		}
	}
}

func ToUpper() Func {
	return func(funcMap template.FuncMap) {
		funcMap["upper"] = strings.ToUpper
	}
}

func ToLower() Func {
	return func(funcMap template.FuncMap) {
		funcMap["lower"] = strings.ToLower
	}
}

func SnakeCase() Func {
	return func(funcMap template.FuncMap) {
		funcMap["snakecase"] = func(s string) string {
			return strcase.ToSnake(s)
		}
	}
}

func CamelCase() Func {
	return func(funcMap template.FuncMap) {
		funcMap["camelcase"] = func(s string) string {
			return strcase.ToCamel(s)
		}
	}
}

func KebabCase() Func {
	return func(funcMap template.FuncMap) {
		funcMap["kebabcase"] = func(s string) string {
			return strcase.ToKebab(s)
		}
	}
}

func Filenamify() Func {
	return func(funcMap template.FuncMap) {
		funcMap["filenamify"] = func(s string) string {
			res, err := filenamify.FilenamifyV2(s)
			if err != nil || res == "" {
				return "invalid-filename"
			}
			return res
		}
	}
}
