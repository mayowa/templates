package templates

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
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

	// register the svg helper when path to svg is provided
	if options.PathToSVG != "" {
		options.FuncMap["svg"] = SvgHelper(options.PathToSVG)
	}

	t.FuncMap = options.FuncMap
	if t.FuncMap == nil {
		t.FuncMap = make(template.FuncMap)
	}

	t.cache = make(map[string]*template.Template)

	t.sharedFolder = filepath.Join(t.root, "shared")
	if err = t.init(); err != nil {
		return nil, err
	}

	t.FuncMap["html"] = func(v string) template.HTML { return template.HTML(v) }
	t.FuncMap["map"] = aMap
	t.FuncMap["slice"] = makeSlice
	t.FuncMap["component"] = t.component
	t.FuncMap["replaceStr"] = replaceStr
	t.FuncMap["ifZero"] = ifZero
	t.FuncMap["attributeSet"] = attributes
	t.FuncMap["deDupeStr"] = deDupeString
	t.FuncMap["mergeTwClasses"] = MergeTwClasses

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

func (t *Template) Exists(name string) bool {
	var (
		found bool
	)

	t.mtx.RLock()
	_, found = t.cache[name]
	t.mtx.RUnlock()

	return found
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

	return err == nil
}

func (t *Template) parse(files ...string) (*template.Template, error) {
	var (
		err      error
		fileList []string
	)

	baseTpl := files[0]
	rfFunc := readFiler(t, t.fSys)

	if t.isFolder(baseTpl) {
		blockFiles, err := t.findFiles(filepath.Join(t.root, baseTpl), t.ext)
		if err != nil {
			return nil, err
		}
		t.sortBlockFiles(baseTpl, blockFiles)

		fileList = append(fileList, blockFiles...)
		if len(files) > 1 {
			fileList = append(fileList, files[1:]...)
		}

		fBaseTpl := baseTpl + "/" + filepath.Base(strings.TrimSuffix(fileList[0], t.ext))
		layout, err := t.extractLayout(fBaseTpl)
		if err != nil && !errors.Is(err, ErrLayoutNotFound) {
			return nil, err
		}

		if layout != "" {
			temp := []string{layout}
			temp = append(temp, fileList...)
			fileList = temp
		}

	} else {
		layout, err := t.extractLayout(baseTpl)
		if err != nil && !errors.Is(err, ErrLayoutNotFound) {
			return nil, err
		}

		// if we have a layout at this point, add it to the fileList
		if layout != "" {
			fileList = append(fileList, layout)
		}

		// add file paths to file list instead of just file names
		for _, fileName := range files {
			fileList = append(fileList, filepath.Join(t.root, fileName+t.ext))
		}
	}

	// parse templates
	tpl, err := t.parseFiles(nil, rfFunc, fileList...)
	if err != nil {
		return nil, err
	}

	// parse shared templates
	filenames, _ := t.findFiles(t.sharedFolder, t.ext)
	if len(filenames) > 0 {
		return t.parseFiles(tpl, rfFunc, filenames...)
	}

	return tpl, nil
}

var extendsRe = regexp.MustCompile(`{{/\*\s*extends?\s*"(.*)"\s*\*/}}`)

var ErrLayoutNotFound = errors.New("layout not found")

func (t *Template) extractLayout(name string) (string, error) {
	if t.isFolder(name) {
		return "", ErrLayoutNotFound
	}

	fle, err := os.Open(filepath.Join(t.root, name+t.ext))
	if err != nil {
		return "", err
	}
	defer fle.Close()

	var src string
	scan := bufio.NewScanner(fle)
	for i := 0; scan.Scan(); i++ {
		src += scan.Text() + "\n"
		if i > 9 {
			break
		}
	}

	match := extendsRe.FindStringSubmatch(src)
	if len(match) < 2 {
		return "", ErrLayoutNotFound
	}

	return filepath.Join(t.root, match[1]+t.ext), nil
}

func (t *Template) sortBlockFiles(blockName string, files []string) {
	// put the file with the same name as the block first
	idx := -1
	for i, fle := range files {
		fle, _ = filepath.Abs(fle)
		fle = strings.TrimSuffix(fle, t.ext)
		fle = filepath.Base(fle)
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

func (t *Template) parseFiles(tpl *template.Template, readFile readFileFunc, filenames ...string) (*template.Template, error) {
	return parseFiles(tpl, readFile, t.FuncMap, filenames)
}

// parseFiles (adapted from stdlib)
func parseFiles(tpl *template.Template, readFile readFileFunc, funcMap template.FuncMap, filenames []string) (*template.Template, error) {
	if len(filenames) == 0 {
		// Not really a problem, but be consistent.
		return nil, fmt.Errorf("html/template: no files named in call to ParseFiles")
	}
	for _, filename := range filenames {
		name, b, err := readFile(filename)
		if err != nil {
			return nil, err
		}

		if err := processComponents(&b); err != nil {
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
			tpl.Funcs(funcMap)
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

	cf := filepath.Join("/", t.componentFolder) + "/"
	if strings.HasPrefix(name, cf) {
		name = strings.TrimPrefix(name, cf)
	} else {
		name = strings.TrimSuffix(name, t.ext)
	}
	if name[0] == '/' {
		name = name[1:]
	}
	return name
}

type readFileFunc func(file string) (name string, b []byte, err error)

// readFile  (adapted from stdlib)
func readFiler(t *Template, fSys fs.FS) readFileFunc {
	return func(file string) (name string, b []byte, err error) {
		if t != nil {
			name = t.stripFileName(file)
		} else {
			name = file
		}

		if fSys != nil {
			b, err = fs.ReadFile(fSys, file)
		} else {
			b, err = os.ReadFile(file)
		}
		return
	}
}
