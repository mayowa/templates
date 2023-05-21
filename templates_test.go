package templates

import (
	"bytes"
	"html/template"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTemplates(t *testing.T) {
	fm := template.FuncMap{
		"upper": strings.ToUpper,
	}

	tpl := NewTemplates("./testData", "tmpl", fm)
	assert.Equal(t, tpl.root, "./testData")
	assert.Equal(t, ".tmpl", tpl.ext)
	assert.Equal(t, "testData/layouts", tpl.layoutFolder)
	assert.Equal(t, "testData/shared", tpl.sharedFolder)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err := tpl.render(buff, "base", "profile", d)
	require.NoError(t, err)
	assert.Equal(t,
		"base layout\n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n",
		buff.String())

}

func Test__noTemplate(t *testing.T) {
	fm := template.FuncMap{
		"upper": strings.ToUpper,
	}
	tpl := NewTemplates("./testData", "tmpl", fm)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err := tpl.render(buff, "", "info", d)
	require.NoError(t, err)
	assert.Equal(t, "make I tell you something...\n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n", buff.String())

}

func Test__templateFolder(t *testing.T) {
	var err error
	fm := template.FuncMap{
		"upper": strings.ToUpper,
	}
	tpl := NewTemplates("./testData", "tmpl", fm)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err = tpl.render(buff, "cast", "block", d)
	require.NoError(t, err)
	assert.Equal(t,
		"cast layout\n\t\t<div>overlay</div>\n    a fragment\n    \n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n\n",
		buff.String())

	buff.Reset()
	err = tpl.render(buff, "", "block", d)
	require.NoError(t, err)
	assert.Equal(t,
		"the index\n\t\t<div>overlay</div>\n    a fragment\n    \n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n\n",
		buff.String())

}
