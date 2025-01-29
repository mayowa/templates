package templates

import (
	"bufio"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

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

func (t *Template) parse(templates ...string) (*template.Template, error) {
	var (
		err      error
		fileList []string
	)

	rfFunc := readFiler(t, t.fSys)
	for i := 0; i < len(templates); i++ {
		tplName := templates[0]
		fls, err := t.getRelatedFiles(tplName)
		if err != nil {
			return nil, err
		}

		fileList = append(fileList, fls...)
	}

	fileList = t.includeLayouts(fileList)

	var tpl *template.Template
	// parse templates
	tpl, err = t.parseFiles(nil, rfFunc, fileList...)
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

func (t *Template) includeLayouts(files []string) []string {
	var fileList []string

	i := 0
	for i < len(files) {
		fileName := files[i]
		layout, err := t.extractLayout(fileName)
		if err == nil && !inFrontOf(files, i, t.cleanTemplateName(layout)) {
			files = slices.Insert(files, i, t.cleanTemplateName(layout))
			continue
		}

		fileList = append(fileList, t.absTemplateName(fileName))
		i++
	}

	return fileList
}

func (t *Template) getRelatedFiles(tpl string) ([]string, error) {
	var (
		fileList []string
	)

	absTpl := t.absTemplateName(tpl)
	fileList = append(fileList, absTpl)

	refs, err := t.findTemplateRefs(absTpl)
	if err != nil {
		return nil, err
	}

	if len(refs) > 0 {
		// remove references to templates not in this folder
		refs = slices.DeleteFunc(refs, func(s string) bool {
			return !t.pathExists(t.absTemplateName(s))
		})

		for _, ref := range refs {
			fileList = append(fileList, t.absTemplateName(ref))
		}
	}

	return fileList, nil
}

var reTplAction = regexp.MustCompile(`{{\-*\s*template\s*"([^"]+)"\s*[\s\w\W]*?\-*}}`)

func (t *Template) findTemplateRefs(file string) ([]string, error) {
	var templates []string

	fileName := t.absTemplateName(file)
	src, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	matches := reTplAction.FindAllStringSubmatch(string(src), -1)
	for _, match := range matches {
		if len(match) > 1 {
			templates = append(templates, match[1])
		}
	}

	return slices.CompactFunc(templates, strings.EqualFold), nil
}

var extendsRe = regexp.MustCompile(`{{/\*\s*extends?\s*"(.*)"\s*\*/}}`)

var ErrLayoutNotFound = errors.New("layout not found")

func (t *Template) extractLayout(name string) (string, error) {
	name = t.cleanTemplateName(name)
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

func (t *Template) cleanTemplateName(name string) string {
	if strings.HasPrefix(name, filepath.Clean(t.root)+string(filepath.Separator)) {
		name = name[len(filepath.Clean(t.root)+string(filepath.Separator)):]
	}

	if strings.HasSuffix(name, t.ext) {
		name = name[:len(name)-len(t.ext)]
	}

	return name
}

func (t *Template) absTemplateName(name string) string {
	name = t.cleanTemplateName(name)
	return filepath.Join(t.root, name+t.ext)
}

func inFrontOf(list []string, idx int, s string) bool {
	if idx == 0 {
		return false
	}

	return list[idx-1] == s
}
