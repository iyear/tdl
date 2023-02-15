package tplfunc

import (
	"text/template"
)

type Func func(funcMap template.FuncMap)

func FuncMap(functions ...Func) template.FuncMap {
	m := make(template.FuncMap)
	for _, f := range functions {
		f(m)
	}
	return m
}

var All []Func

func init() {
	mods := [][]Func{String, Math, Date}
	for _, mod := range mods {
		All = append(All, mod...)
	}
}
