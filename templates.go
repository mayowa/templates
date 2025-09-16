package templates

import (
	"bytes"
	"errors"
	"fmt"
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
	parsing            map[string]*sync.Mutex
	mtx                sync.RWMutex
	debug              bool
	fSys               fs.FS
	componentFolder    string
	componentTemplates *template.Template
}

type TemplateOptions struct {
	Ext       string
	FuncMap   template.FuncMap
	PathToSVG string
	FS        fs.FS
	Debug     bool
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
	t.debug = options.Debug

	// default to .tmpl when none is provided
	if options.Ext == "" {
		options.Ext = ".tmpl"
	}

	t.ext = options.Ext
	if options.Ext[0] != '.' {
		t.ext = "." + options.Ext
	}

	t.cache = make(map[string]*template.Template)
	t.parsing = make(map[string]*sync.Mutex)

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

type RenderOption struct {
	Layout       string
	Template     string
	RenderString bool
	Others       []string
	Data         any
}

func (t *Template) Render(out io.Writer, option RenderOption) error {
	return t.renderFiles(out, option.Layout, option.Template, option.Data, option.Others)
}

var ErrNoTemplates = errors.New("no templates")

func (t *Template) renderFiles(out io.Writer, layout, name string, data any, others []string) error {
	var (
		err   error
		found bool
		tpl   *template.Template
	)

	if layout == "" && name == "" {
		return ErrNoTemplates
	}

	var templates []string

	baseTpl := fmt.Sprint("noLayout", "-", name)
	if layout != "" {
		baseTpl = fmt.Sprint(layout, "-", name)
	}

	var parseLock *sync.Mutex
	var plFound bool
	t.mtx.Lock()
	parseLock, plFound = t.parsing[baseTpl]
	if !plFound {
		t.parsing[baseTpl] = &sync.Mutex{}
		parseLock = t.parsing[baseTpl]
	}
	t.mtx.Unlock()

	parseLock.Lock()
	if !t.debug {
		t.mtx.RLock()
		tpl, found = t.cache[baseTpl]
		t.mtx.RUnlock()
	}

	if !found {
		// put layout first if provided
		// expand the first entry in templates if it includes multiple files
		templates = append([]string{name}, others...)

		if layout != "" {
			templates = append([]string{layout}, templates...)
		}

		tpl, err = t.parse(templates...)
		if err != nil {
			parseLock.Unlock()
			return err
		}

		if !t.debug {
			t.mtx.Lock()
			t.cache[baseTpl] = tpl
			t.mtx.Unlock()
		}

	}
	parseLock.Unlock()

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
