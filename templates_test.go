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

var options = &TemplateOptions{
	Ext:       "tmpl",
	FuncMap:   fm,
	PathToSVG: "./testData/svg",
}

func TestNewTemplates(t *testing.T) {

	tpl, err := New("./testData", options)
	require.NoError(t, err)
	assert.Equal(t, tpl.root, "./testData")
	assert.Equal(t, ".tmpl", tpl.ext)
	assert.Equal(t, "testData/shared", tpl.sharedFolder)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err = tpl.Render(buff, RenderOption{Layout: "", Template: "profile", Data: d})
	require.NoError(t, err)
	assert.Equal(t,
		"cast layout\n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n",
		buff.String(),
	)

}

func Test_templateCache(t *testing.T) {

	tpl, err := New("./testData", options)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err = tpl.Render(buff, RenderOption{Layout: "", Template: "profile", Data: d})
	require.NoError(t, err)
	assert.Contains(t, tpl.cache, "noLayout-profile")

	tpl, err = New("./testData", options)
	require.NoError(t, err)

	tpl.Debug = true
	err = tpl.Render(buff, RenderOption{Layout: "", Template: "profile", Data: d})
	require.NoError(t, err)
	assert.Contains(t, tpl.cache, "noLayout-profile")
}

func Test__noTemplate(t *testing.T) {

	tpl, err := New("./testData", options)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err = tpl.Render(buff, RenderOption{Layout: "", Template: "info", Data: d})

	require.NoError(t, err)

	require.NoError(t, err)
	assert.Equal(t, "make I tell you something...\n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n", buff.String())

}

func Test__nestedExtends(t *testing.T) {
	var err error

	tpl, err := New("./testData", options)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)

	err = tpl.Render(buff, RenderOption{Layout: "", Template: "child", Data: nil})

	require.NoError(t, err)
	assert.Equal(t,
		"i'm the grandpai'm the dadi'm the child",
		buff.String())
}

func TestStringWithLayout(t *testing.T) {
	var err error
	tpl, err := New("./testData", options)
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
	tpl, err := New("./testData", options)
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

	tpl, err := New("./testData", options)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)
	err = tpl.Render(buff, RenderOption{Layout: "", Template: "inFolder/index", Data: nil})

	require.NoError(t, err)
	assert.True(t, tpl.InCache("", "inFolder/index"))
}

func TestTemplate_NoShared(t *testing.T) {
	var err error

	tpl, err := New("./testData", options)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}

	err = tpl.Render(buff, RenderOption{Layout: "", Template: "solo", Data: d})

	require.NoError(t, err)
	assert.Equal(t, "philippta, This is solo act!", buff.String())
}

func Test_Components(t *testing.T) {
	tpl, err := New("./testData", options)
	require.NoError(t, err)

	buff := bytes.NewBuffer(nil)
	err = tpl.Render(buff, RenderOption{Layout: "", Template: "comp-demo", Data: nil})

	require.NoError(t, err)
	assert.Equal(t, buff.String(), "<div>\n    \n    <div class=\"isCard\">\n\t<h1>this cards title</h1>\n\t<p>its a brand new day</p>\n    <div class=\"isCard\">\n\t<h1>a that is self enclosed and nested</h1>\n</div>\n\t<h2>Another one?</h2>\n    <div class=\"isCard\">\n\t<h1>nested dolls...</h1>\n\there we come.... wait are we russian??\n    \n</div>\n    \n</div>\n</div>")
}

func Test_ComponentRenderer(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	tpl, err := New("./testData", options)
	require.NoError(t, err)

	if err != nil {
		t.Fatal(err.Error())
	}

	err = tpl.Render(buff, RenderOption{Layout: "", Template: "comp-renderer", Data: nil})

	require.NoError(t, err)
	output := buff.String()
	t.Log(output)
	assert.Equal(t, "\n<div class=\"isBox\">\n\t<h1 class=\"bar?\">this cards title</h1>\n\tliving large within a Box!!\n\t<div class=\"isCard\">\n\t<h1>Ode to a card</h1>A card within a box?\n</div>\n</div>\n", output)
}

func Test_ComplexComponentParams(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	tpl, err := New("./testData", options)
	require.NoError(t, err)

	type SelectOption struct {
		Label    string
		Value    string
		Selected bool
		Disabled bool
	}

	opts := []SelectOption{
		{"Pick something", "0", true, true},
		{"Male", "1", false, false},
		{"Female", "2", false, false},
	}

	data := map[string]any{
		"Select": opts,
	}

	err = tpl.Render(buff, RenderOption{Layout: "", Template: "complex-params", Data: data})

	require.NoError(t, err)
	output := buff.String()
	t.Log(output)
	assert.Equal(t, "<select \n\t\tname=\"\" \n\t\tid=\"\" \n\t\tclass=\"bg-gray-50 border border-gray-300 text-content text-sm rounded-lg p-2.5 &lt;nil&gt;\"\n\t><option value=\"0\" \n\t\t\t selected  \n\t\t\t disabled  \n\t\t>\n\t\t\tPick something\n\t\t</option><option value=\"1\" \n\t\t\t \n\t\t\t \n\t\t>\n\t\t\tMale\n\t\t</option><option value=\"2\" \n\t\t\t \n\t\t\t \n\t\t>\n\t\t\tFemale\n\t\t</option></select>\n\n",
		output)
}

func Test__PassParamsToEnd(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	tpl, err := New("./testData", options)
	require.NoError(t, err)

	data := map[string]string{
		"name": "Paul",
	}

	err = tpl.Render(buff, RenderOption{Layout: "", Template: "p-to-end", Data: data})

	require.NoError(t, err)
	output := buff.String()

	assert.Equal(t, "<div class=\"isHi\">\n\t<p> Hi, Paul! </p>\n\t<p> Hello, Paul! </p>\n</div>", output)
}

func Test__NestedComponents(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	tpl, err := New("./testData", options)
	require.NoError(t, err)

	if err != nil {
		t.Fatal(err.Error())
	}

	err = tpl.Render(buff, RenderOption{Layout: "", Template: "comp-dialog", Data: nil})

	require.NoError(t, err)
	output := buff.String()
	t.Log(output)
	assert.Equal(t, "\n<div class=\"isDialog\">\n\t<div class=\"isBox\">\n\t<h1 class=\"\">this box title</h1>\n\ta box living large within a Dialog!\n\t<div class=\"isCard\">\n\t<h1>Ode to a box</h1>A card within a box?\n</div>\n</div>\n\t<button>OK</button>\n</div>\n", output)
}

func Test__InFolder(t *testing.T) {
	buff := bytes.NewBuffer(nil)
	tpl, err := New("./testData", options)
	require.NoError(t, err)

	tests := []struct {
		name     string
		tplName  string
		expected string
	}{
		{name: "with template in shared", tplName: "inFolderWithShared/index", expected: "A template\nI am a form"},
		{name: "no block overide", tplName: "inFolder/index", expected: "A template\nwith no content"},
		{name: "extends outside template", tplName: "inFolder/child", expected: "base layout\nA child"},
		{name: "nested extend", tplName: "inFolder/grandchild", expected: "base layout\nA child and grand child"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buff.Reset()
			err = tpl.Render(buff, RenderOption{Layout: "", Template: tt.tplName, Data: nil})

			require.NoError(t, err)
			output := buff.String()
			assert.Equal(t, tt.expected, output)

		})
	}
}

func Test_SpecifyDyanmicLayout(t *testing.T) {
	tpl, err := New("./testData", options)
	require.NoError(t, err)
	assert.Equal(t, tpl.root, "./testData")
	assert.Equal(t, ".tmpl", tpl.ext)
	assert.Equal(t, "testData/shared", tpl.sharedFolder)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err = tpl.Render(buff, RenderOption{Layout: "cast", Template: "multi", Data: d})

	require.NoError(t, err)
	assert.Equal(t,
		"cast layout\n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n",
		buff.String(),
	)
}
