package templates

import (
	"bytes"
	"html/template"
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
