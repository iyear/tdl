package texpr

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/fatih/color"
)

type FieldsGetter struct {
	opts *Options
}

type Field struct {
	Path    []string
	Type    reflect.Type
	Comment string
}

type Options struct {
	tagName string
}

type Option func(opts *Options)

func NewFieldsGetter(opts *Options) *FieldsGetter {
	if opts == nil {
		opts = &Options{
			tagName: "comment",
		}
	}

	return &FieldsGetter{
		opts: opts,
	}
}

func (f *FieldsGetter) Sprint(fields []*Field, colorable bool) string {
	b := &strings.Builder{}

	for _, field := range fields {
		path := strings.Join(field.Path, ".")
		if colorable {
			path = color.BlueString(path)
		}

		typ := field.Type.String()
		if colorable {
			typ = color.GreenString(typ)
		}

		comment := "# " + field.Comment
		if colorable {
			comment = color.MagentaString(comment)
		}

		b.WriteString(fmt.Sprintf("%s: %s %s\n", path, typ, comment))
	}

	return b.String()
}

func (f *FieldsGetter) Walk(v any) ([]*Field, error) {
	value := reflect.TypeOf(v)
	if value.Kind() != reflect.Struct && value.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("please input a struct")
	}

	fields := make([]*Field, 0)
	f.walk(value, &Field{}, &fields)

	return fields, nil
}

func (f *FieldsGetter) walk(v reflect.Type, field *Field, fields *[]*Field) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String, reflect.Bool, reflect.Map:

		field.Type = v
		*fields = append(*fields, field)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			fd := v.Field(i)

			if !fd.IsExported() {
				continue
			}

			f.walk(fd.Type, &Field{
				Path:    append(field.Path, fd.Name),
				Comment: fd.Tag.Get(f.opts.tagName),
			}, fields)
		}
	case reflect.Array, reflect.Slice:
		field.Path[len(field.Path)-1] += "[]" // note this is an array or slice
		f.walk(v.Elem(), field, fields)
	case reflect.Pointer:
		f.walk(v.Elem(), field, fields)
	default:
		// do nothing
	}
}
