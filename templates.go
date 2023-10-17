package templates

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Template struct {
	root         string
	ext          string
	sharedFolder string
	FuncMap      template.FuncMap

	cache      map[string]*template.Template
	mtx        sync.RWMutex
	Debug      bool
	components *template.Template
}

func New(root, ext string, funcMap template.FuncMap) (*Template, error) {
	var err error
	t := new(Template)
	t.root = root
	t.ext = ext
	if ext[0] != '.' {
		t.ext = "." + ext
	}
	t.FuncMap = funcMap
	t.cache = make(map[string]*template.Template)

	t.sharedFolder = filepath.Join(t.root, "shared")
	if err := t.init(); err != nil {
		return nil, err
	}

	t.FuncMap["html"] = func(v string) template.HTML { return template.HTML(v) }

	// components templates
	componentsFolder := "components"
	if t.isFolder(componentsFolder) {
		t.components, err = template.New("").Funcs(t.FuncMap).ParseGlob(filepath.Join(t.root, componentsFolder) + "/*" + t.ext)
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}

func (t *Template) init() error {
	return nil
}

func (t *Template) Render(out io.Writer, layout, name string, data any) error {
	return t.RenderFiles(out, data, layout, name)
}

func (t *Template) RenderFiles(out io.Writer, data any, layout, name string, others ...string) error {
	var (
		err   error
		found bool
		tpl   *template.Template
	)

	if !t.Debug {
		t.mtx.RLock()
		tpl, found = t.cache[name]
		t.mtx.RUnlock()
	}

	if !found {
		others = append([]string{name}, others...)
		tpl, err = t.parse(layout, others...)
		if err != nil {
			return err
		}

		t.mtx.Lock()
		t.cache[name] = tpl
		t.mtx.Unlock()
	}

	return tpl.Execute(out, data)
}
func (t *Template) Exists(name string) bool {
	var (
		found bool
	)

	t.mtx.RLock()
	_, found = t.cache[name]
	t.mtx.RUnlock()
	if found {
		return true
	}

	return false
}

func (t *Template) String(layout, src string, data any) (string, error) {
	var (
		err error
		tpl *template.Template
	)

	if layout != "" {
		layoutFleName := filepath.Join(t.root, layout+t.ext)
		tpl, err = t.parseFiles(nil, t.readFileOS, layoutFleName)
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
		if tpl, err = t.parseFiles(tpl, t.readFileOS, filenames...); err != nil {
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

// isFolder checks if a folder exists in the template folder
func (t *Template) isFolder(name string) bool {
	templateName := filepath.Join(t.root, name)
	fi, err := os.Stat(templateName)
	if err != nil {
		return false
	}

	return fi.Mode().IsDir()
}

func (t *Template) pathExists(name string) bool {
	_, err := os.Stat(name)
	if err != nil {
		return false
	}

	return true
}

func (t *Template) parse(layout string, names ...string) (*template.Template, error) {
	if len(names) == 0 {
		return nil, errors.New("templates not specified")
	}

	name := names[0]
	if len(names) > 1 {
		names = names[1:]
	}

	layoutFleName := filepath.Join(t.root, layout+t.ext)
	templateName := filepath.Join(t.root, name+t.ext)

	var files []string
	if layout != "" {
		files = append(files, layoutFleName)
	}

	if t.isFolder(name) {
		filenames, err := t.findFiles(filepath.Join(t.root, name), t.ext)
		if err != nil {
			return nil, err
		}
		t.sortBlockFiles(name, filenames)
		files = append(files, filenames...)
	} else {
		files = append(files, templateName)
	}

	tpl, err := t.parseFiles(nil, t.readFileOS, files...)
	if err != nil {
		return nil, err
	}
	filenames, _ := t.findFiles(t.sharedFolder, t.ext)
	if len(filenames) > 0 {
		return t.parseFiles(tpl, t.readFileOS, filenames...)
	}

	return tpl, nil
}

func (t *Template) sortBlockFiles(blockName string, files []string) {
	// put the file with the same name as the block first
	idx := -1
	for i, fle := range files {
		fle, _ = filepath.Abs(fle)
		fle = strings.TrimSuffix(fle, t.ext)
		if fle == blockName {
			idx = i
			break
		}
	}

	if idx == -1 {
		return
	}
	files[0], files[idx] = files[idx], files[0]
}

// findFiles
func (t *Template) findFiles(root, ext string) (filenames []string, err error) {

	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		} else if d.IsDir() {
			return nil
		}

		if strings.HasSuffix(path, ext) {
			filenames = append(filenames, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(filenames) == 0 {
		return nil, fmt.Errorf("no file found in: %#q", root)
	}

	return filenames, nil
}

type readFileFunc func(string) (string, []byte, error)

// parseFiles (adapted from stdlib)
func (t *Template) parseFiles(tpl *template.Template, readFile readFileFunc, filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("html/template: no files named in call to ParseFiles")
	}
	for _, filename := range filenames {
		name, b, err := readFile(filename)
		if err != nil {
			return nil, err
		}

		if err := t.processComponentsInTemplate(&b); err != nil {
			return nil, err
		}

		s := string(b)
		// First template becomes return value if not already defined,
		// and we use that one for subsequent New calls to associate
		// all the templates together. Also, if this file has the same name
		// as t, this file becomes the contents of t, so
		//  t, err := New(name).Funcs(xxx).ParseFiles(name)
		// works. Otherwise we create a new template associated with t.
		var tmpl *template.Template
		if tpl == nil {
			tpl = template.New(name)
			tpl.Funcs(t.FuncMap)
		}
		if name == tpl.Name() {
			tmpl = tpl
		} else {
			tmpl = tpl.New(name)
		}
		_, err = tmpl.Parse(s)
		if err != nil {
			return nil, err
		}
	}
	return tpl, nil
}

func (t *Template) stripFileName(name string) string {
	if strings.HasPrefix(name, t.sharedFolder) {
		name = strings.TrimPrefix(name, t.sharedFolder)
	} else {
		name = strings.TrimPrefix(name, filepath.Clean(t.root))
	}
	name = strings.TrimSuffix(name, t.ext)

	if name[0] == '/' {
		name = name[1:]
	}
	return name
}

// readFileOS  (adapted from stdlib)
func (t *Template) readFileOS(file string) (name string, b []byte, err error) {
	name = t.stripFileName(file)
	b, err = os.ReadFile(file)
	return
}

// readFileFS  (borrowed from stdlib)
func (t *Template) readFileFS(fsys fs.FS) func(string) (string, []byte, error) {
	return func(file string) (name string, b []byte, err error) {
		name = t.stripFileName(file)
		b, err = fs.ReadFile(fsys, file)
		return
	}
}
