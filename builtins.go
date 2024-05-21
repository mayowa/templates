package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"
	"strings"
)

func (t *Template) component(name string, args map[any]any) template.HTML {
	name += t.ext
	tpl := t.componentTemplates.Lookup(name)
	if tpl == nil {
		return ""
	}

	buff := bytes.NewBufferString("")
	if err := tpl.Execute(buff, args); err != nil {
		return template.HTML(err.Error())
	}

	return template.HTML(buff.String())
}

func aMap(args ...any) map[any]interface{} {
	retv := make(map[any]interface{}, len(args))
	for i := 0; i < len(args); i += 2 {
		retv[args[i]] = args[i+1]
	}

	return retv
}

func ifZero(src any, def any) any {
	if src == nil {
		return ""
	}

	vs := reflect.ValueOf(src)
	vs = reflect.Indirect(vs)

	if vs.IsZero() {
		return def
	}

	return src
}

func replaceStr(str, old, new string) string {
	return strings.Replace(str, old, new, -1)
}

type HTMLAttributes map[string]string

func attributes() *HTMLAttributes {
	a := make(HTMLAttributes)
	return &a
}

func (a *HTMLAttributes) Set(key, value string) string {
	(*a)[key] = value
	return ""
}

func (a *HTMLAttributes) Render() template.HTMLAttr {
	out := ""
	for k, v := range *a {
		out += fmt.Sprintf("%s=%q ", k, v)
	}

	return template.HTMLAttr(out)
}
