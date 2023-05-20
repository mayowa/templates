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

	tpl := NewTemplates("./tplfiles", "tmpl", fm)
	assert.Equal(t, tpl.root, "./tplfiles")
	assert.Equal(t, ".tmpl", tpl.ext)
	assert.Equal(t, "tplfiles/layouts", tpl.layoutFolder)
	assert.Equal(t, "tplfiles/shared", tpl.sharedFolder)

	buff := bytes.NewBuffer(nil)
	d := struct{ Name string }{Name: "philippta"}
	err := tpl.render(buff, "base", "profile", d)
	require.NoError(t, err)
	assert.Equal(t,
		"base layout\n<div class=\"profile\">\n  Your username: PHILIPPTA\n</div>\n",
		buff.String())

}
