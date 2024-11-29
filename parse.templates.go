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
			fileList = append([]string{layout}, fileList...)
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
