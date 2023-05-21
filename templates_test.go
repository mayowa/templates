package templates

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fm = template.FuncMap{
	"upper": strings.ToUpper,
}

func TestNewTemplates(t *testing.T) {

	tpl := NewTemplates("./testData", "tmpl", fm)
	assert.Equal(t, tpl.root, "./testData")
	assert.Equal(t, ".tmpl", tpl.ext)
	assert.Equal(t, "testData/layouts", tpl.layoutFolder)
	assert.Equal(t, "testData/shared", tpl.sharedFolder)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err := tpl.Render(buff, "base", "profile", d)
	require.NoError(t, err)
	assert.Equal(t,
		"base layout\n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n",
		buff.String())

}

func Test_templateCache(t *testing.T) {

	tpl := NewTemplates("./testData", "tmpl", fm)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err := tpl.Render(buff, "base", "profile", d)
	require.NoError(t, err)
	assert.Contains(t, tpl.cache, "base-profile")

	tpl = NewTemplates("./testData", "tmpl", fm)
	tpl.Debug = true
	err = tpl.Render(buff, "base", "profile", d)
	require.NoError(t, err)
	assert.NotContains(t, tpl.cache, "base-profile")
}

func Test__noTemplate(t *testing.T) {

	tpl := NewTemplates("./testData", "tmpl", fm)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err := tpl.Render(buff, "", "info", d)
	require.NoError(t, err)
	assert.Equal(t, "make I tell you something...\n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n", buff.String())

}

func Test__templateFolder(t *testing.T) {
	var err error

	tpl := NewTemplates("./testData", "tmpl", fm)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err = tpl.Render(buff, "cast", "block", d)
	require.NoError(t, err)
	assert.Equal(t,
		"cast layout\n\t\t<div>overlay</div>\n    a fragment\n    \n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n\n",
		buff.String())

	buff.Reset()
	err = tpl.Render(buff, "", "block", d)
	require.NoError(t, err)
	assert.Equal(t,
		"the index\n\t\t<div>overlay</div>\n    a fragment\n    \n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n\n",
		buff.String())

}

func TestStringWithLayout(t *testing.T) {
	var err error
	tpl := NewTemplates("./testData", "tmpl", fm)

	out := ""
	d := struct{ Name string }{Name: "philippta"}
	out, err = tpl.String("string", `
		{{define "string"}}
			a string block
			{{template "modal/overlay"}}
		{{end}}
	`, d)
	require.NoError(t, err)
	assert.Equal(t,
		"string layout\n\t\t\ta string block\n\t\t\tYou cant see me!\n\t\t",
		out)

}
func TestStringWithoutLayout(t *testing.T) {
	var err error
	tpl := NewTemplates("./testData", "tmpl", fm)

	out := ""
	d := struct{ Name string }{Name: "philippta"}
	out, err = tpl.String("", `
		the index
		{{- block "content" .}}
			a string block
		{{end -}}
	`, d)
	require.NoError(t, err)
	assert.Equal(t,
		"\n\t\tthe index\n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n",
		out)

}
