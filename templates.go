package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Templates struct {
	root         string
	ext          string
	layoutFolder string
	sharedFolder string
	funcMap      template.FuncMap

	cache map[string]*template.Template
	mtx   sync.RWMutex
	Debug bool
}

func NewTemplates(root, ext string, funcMap template.FuncMap) *Templates {
	t := new(Templates)
	t.root = root
	t.ext = ext
	if ext[0] != '.' {
		t.ext = "." + ext
	}
	t.funcMap = funcMap
	t.cache = make(map[string]*template.Template)

	t.layoutFolder = filepath.Join(t.root, "layouts")
	t.sharedFolder = filepath.Join(t.root, "shared")
	return t
}

func (t *Templates) Render(out io.Writer, layout, name string, data any) error {
	var (
		err   error
		found bool
		tpl   *template.Template
	)

	if !t.Debug {
		t.mtx.RLock()
		tpl, found = t.cache[layout+"-"+name]
		t.mtx.RUnlock()
	}

	if !found {
		tpl, err = t.parse(layout, name)
		if err != nil {
			return err
		}

		if !t.Debug {
			t.mtx.Lock()
			t.cache[layout+"-"+name] = tpl
			t.mtx.Unlock()
		}
	}

	return tpl.Execute(out, data)
}

func (t *Templates) String(layout, src string, data any) (string, error) {
	var (
		err error
		tpl *template.Template
	)

	if layout != "" {
		layoutFleName := filepath.Join(t.layoutFolder, layout+t.ext)
		tpl, err = t.parseFiles(nil, t.readFileOS, layoutFleName)
		if err != nil {
			return "", err
		}

		tpl, err = tpl.Parse(src)
		if err != nil {
			return "", err
		}
	} else {
		tpl, err = template.New("").Funcs(t.funcMap).Parse(src)
		if err != nil {
			return "", err
		}
	}

	filenames, err := t.findFiles(t.sharedFolder, t.ext)
	if err != nil {
		return "", err
	}

	tpl, err = t.parseFiles(tpl, t.readFileOS, filenames...)
	if err != nil {
		return "", err
	}

	out := bytes.NewBufferString("")
	err = tpl.Execute(out, data)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func (t *Templates) isFolder(name string) bool {
	templateName := filepath.Join(t.root, name)
	fi, err := os.Stat(templateName)
	if err != nil {
		return false
	}

	return fi.Mode().IsDir()
}

func (t *Templates) parse(layout, name string) (*template.Template, error) {
	layoutFleName := filepath.Join(t.layoutFolder, layout+t.ext)
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
	filenames, err := t.findFiles(t.sharedFolder, t.ext)
	if err != nil {
		return nil, err
	}

	return t.parseFiles(tpl, t.readFileOS, filenames...)
}

func (t *Templates) sortBlockFiles(blockName string, files []string) {
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
func (t *Templates) findFiles(root, ext string) (filenames []string, err error) {

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
func (t *Templates) parseFiles(tpl *template.Template, readFile readFileFunc, filenames ...string) (*template.Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("html/template: no files named in call to ParseFiles")
	}
	for _, filename := range filenames {
		name, b, err := readFile(filename)

		if err != nil {
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
			tpl.Funcs(t.funcMap)
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

func (t *Templates) stripFileName(name string) string {
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
func (t *Templates) readFileOS(file string) (name string, b []byte, err error) {
	name = t.stripFileName(file)
	b, err = os.ReadFile(file)
	return
}

// readFileFS  (borrowed from stdlib)
func (t *Templates) readFileFS(fsys fs.FS) func(string) (string, []byte, error) {
	return func(file string) (name string, b []byte, err error) {
		name = t.stripFileName(file)
		b, err = fs.ReadFile(fsys, file)
		return
	}
}
