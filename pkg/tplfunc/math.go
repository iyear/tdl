package tplfunc

import (
	"math/rand"
	"text/template"
	"time"
)

var Math = []Func{Rand()}

func init() {
	rand.Seed(time.Now().Unix())
}

func Rand() Func {
	return func(funcMap template.FuncMap) {
		funcMap["rand"] = func(min, max int) int {
			return rand.Intn(max-min) + min
		}
	}
}
