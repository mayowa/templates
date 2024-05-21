package templates

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fm = template.FuncMap{
	"upper": strings.ToUpper,
}

func TestNewTemplates(t *testing.T) {

	tpl, err := New("./testData", "tmpl", fm)
	require.NoError(t, err)
	assert.Equal(t, tpl.root, "./testData")
	assert.Equal(t, ".tmpl", tpl.ext)
	assert.Equal(t, "testData/shared", tpl.sharedFolder)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err = tpl.Render(buff, "profile", d)
	require.NoError(t, err)
	assert.Nil(t, deep.Equal(
		"cast layout\n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n",
		buff.String()))

}

func Test_templateCache(t *testing.T) {

	tpl, err := New("./testData", "tmpl", fm)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err = tpl.Render(buff, "profile", d)
	require.NoError(t, err)
	assert.Contains(t, tpl.cache, "profile")

	tpl, err = New("./testData", "tmpl", fm)
	require.NoError(t, err)

	tpl.Debug = true
	err = tpl.Render(buff, "profile", d)
	require.NoError(t, err)
	assert.Contains(t, tpl.cache, "profile")
}

func Test__noTemplate(t *testing.T) {

	tpl, err := New("./testData", "tmpl", fm)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err = tpl.Render(buff, "info", d)
	require.NoError(t, err)

	require.NoError(t, err)
	assert.Equal(t, "make I tell you something...\n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n", buff.String())

}

func Test__templateFolder(t *testing.T) {
	var err error

	tpl, err := New("./testData", "tmpl", fm)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err = tpl.Render(buff, "block", d)
	require.NoError(t, err)
	assert.Equal(t,
		"the index\n\t\t<div>overlay</div>\n    a fragment\n    \n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n\n",
		buff.String())
}

func Test__templateInsideAFolder(t *testing.T) {
	var err error

	tpl, err := New("./testData", "tmpl", fm)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)
	err = tpl.Render(buff, "block/fragment", nil)
	require.NoError(t, err)
	assert.Equal(t,
		"a fragment", buff.String())

}

func TestStringWithLayout(t *testing.T) {
	var err error
	tpl, err := New("./testData", "tmpl", fm)
	require.NoError(t, err)

	out := ""
	d := struct{ Name string }{Name: "philippta"}
	out, err = tpl.String("string", `
	
		{{define "another"}}one{{end}}
	
		{{define "main"}}
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
	tpl, err := New("./testData", "tmpl", fm)
	require.NoError(t, err)

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

func TestTemplate_Lookup(t *testing.T) {
	var err error

	tpl, err := New("./testData", "tmpl", fm)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err = tpl.Render(buff, "block", d)
	require.NoError(t, err)
	assert.True(t, tpl.Exists("block"))
}

func TestTemplate_NoShared(t *testing.T) {
	var err error

	tpl, err := New("./testData", "tmpl", fm)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err = tpl.Render(buff, "solo", d)
	require.NoError(t, err)
	assert.Equal(t, "philippta, This is solo act!", buff.String())
}

func Test_Components(t *testing.T) {
	tpl, err := New("./testData", "tmpl", fm)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)
	err = tpl.Render(buff, "comp-demo", nil)
	require.NoError(t, err)
	assert.Equal(t, buff.String(), "<div>\n    \n    <div class=\"isCard\">\n\t<h1>this cards title</h1>\n\t<p>its a brand new day</p>\n    <div class=\"isCard\">\n\t<h1>a that is self enclosed and nested</h1></div>\n\t<h2>Another one?</h2>\n    <div class=\"isCard\">\n\t<h1>nested dolls...</h1>\n\there we come.... wait are we russian??\n    </div>\n    </div>\n</div>")
}

func Test_ComponentRenderer(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	tpl, err := New("./testData", "tmpl", fm)
	require.NoError(t, err)

	if err != nil {
		t.Fatal(err.Error())
	}

	err = tpl.Render(buff, "comp-renderer", nil)
	require.NoError(t, err)
	output := buff.String()
	t.Log(output)
	assert.Equal(t, output, "\n<div class=\"isBox\">\n\t<h1 class=\"bar?\">this cards title</h1>\n\tliving large within a Box!!\n\t<div class=\"isCard\">\n\t<h1>Ode to a card</h1>A card within a box?\n</div>\n</div>\n")
}
