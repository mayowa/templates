package templates

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"io/fs"
	"path/filepath"
	"strings"
	"sync"
)

type Template struct {
	root         string
	ext          string
	sharedFolder string
	FuncMap      template.FuncMap

	cache              map[string]*template.Template
	mtx                sync.RWMutex
	Debug              bool
	fSys               fs.FS
	componentFolder    string
	componentTemplates *template.Template
}

type TemplateOptions struct {
	Ext       string
	FuncMap   template.FuncMap
	PathToSVG string
	FS        fs.FS
}

func New(root string, options *TemplateOptions) (*Template, error) {
	var err error
	t := new(Template)
	t.root = root

	if options == nil {
		options = &TemplateOptions{
			Ext:       ".tmpl",
			FuncMap:   template.FuncMap{},
			PathToSVG: "./resources/svg",
		}
	}

	t.fSys = options.FS

	// default to .tmpl when none is provided
	if options.Ext == "" {
		options.Ext = ".tmpl"
	}

	t.ext = options.Ext
	if options.Ext[0] != '.' {
		t.ext = "." + options.Ext
	}

	t.cache = make(map[string]*template.Template)

	t.sharedFolder = filepath.Join(t.root, "shared")
	if err = t.init(); err != nil {
		return nil, err
	}

	if err = BuiltinInit(t, options); err != nil {
		return nil, err
	}

	// components templates
	t.componentFolder = "components"
	if t.isFolder(t.componentFolder) {
		componentFolder := filepath.Join(t.root, t.componentFolder)
		t.componentTemplates, err = componentTemplates(componentFolder, t.ext, t.FuncMap, readFiler(t, t.fSys))
		if err != nil && !strings.Contains(err.Error(), "pattern matches no files") {
			return nil, err
		}
	}
	return t, nil
}

func (t *Template) init() error {
	return nil
}

func (t *Template) Render(out io.Writer, name string, data any) error {
	return t.RenderFiles(out, data, name)
}

var ErrNoTemplates = errors.New("no templates")

func (t *Template) RenderFiles(out io.Writer, data any, templates ...string) error {
	var (
		err   error
		found bool
		tpl   *template.Template
	)

	if len(templates) == 0 {
		return ErrNoTemplates
	}

	baseTpl := templates[0]

	if !t.Debug {
		t.mtx.RLock()
		tpl, found = t.cache[baseTpl]
		t.mtx.RUnlock()
	}

	if !found {
		tpl, err = t.parse(templates...)
		if err != nil {
			return err
		}

		t.mtx.Lock()
		t.cache[baseTpl] = tpl
		t.mtx.Unlock()
	}

	return tpl.Execute(out, data)
}

func (t *Template) String(layout, src string, data any) (string, error) {
	var (
		err error
		tpl *template.Template
	)

	if layout != "" {
		layoutFleName := filepath.Join(t.root, layout+t.ext)
		tpl, err = t.parseFiles(nil, readFiler(t, t.fSys), layoutFleName)
		if err != nil {
			return "", err
		}

		tpl, err = tpl.Parse(src)
		if err != nil {
			return "", err
		}
	} else {
		tpl, err = template.New("").Funcs(t.FuncMap).Parse(src)
		if err != nil {
			return "", err
		}
	}

	filenames, _ := t.findFiles(t.sharedFolder, t.ext)
	if err != nil {
		return "", err
	}

	if len(filenames) > 0 {
		if tpl, err = t.parseFiles(tpl, readFiler(t, t.fSys), filenames...); err != nil {
			return "", err
		}
	}

	out := bytes.NewBufferString("")
	err = tpl.Execute(out, data)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func (t *Template) getLayout() string {
	return ""
}
