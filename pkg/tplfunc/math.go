package tplfunc

import (
	"math/rand"
	"text/template"
	"time"
)

var Math = []Func{Rand()}

var rnd *rand.Rand

func init() {
	rnd = rand.New(rand.NewSource(time.Now().Unix()))
}

func Rand() Func {
	return func(funcMap template.FuncMap) {
		funcMap["rand"] = func(min, max int) int {
			return rnd.Intn(max-min) + min
		}
	}
}
