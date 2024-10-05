package templates

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"path/filepath"
	"strings"
)

func (t *Template) processComponentsInTemplate(contents *[]byte) error {

	if t.componentTemplates == nil {
		return nil
	}

	return processComponents(contents)
}

func processComponents(contents *[]byte) error {
	buf := bytes.NewBuffer(nil)
	var tags []*Tag
	for {
		cTag, err := findNextTag(*contents)
		if err != nil {
			return err
		}
		if cTag == nil {
			break
		}

		if cTag.IsEnd {
			startTag := findStartTag(cTag, tags)
			if startTag == nil {
				return errors.New("template: unable to find start tag for " + cTag.Name)
			}

			cTag.Args = startTag.Args
		}
		tags = append(tags, cTag)

		cName := strings.ToLower(cTag.Name)
		args := fmt.Sprintf(`(map "_isSelfClosing" %v "_isEnd" %v %s)`, cTag.IsSelfClosing, cTag.IsEnd, cTag.Args.ArgPairs())
		buf.WriteString(`{{ component "` + cName + `" ` + args + ` }}`)

		// replace rendered component with tag block
		start := cTag.loc[0]
		end := cTag.loc[1]
		*contents = append((*contents)[:start], append(buf.Bytes(), (*contents)[end:]...)...)
		// tHalf := (*contents)[:start]
		// bHalf := (*contents)[end:]
		// *contents = append(tHalf, append(buf.Bytes(), bHalf...)...)
		buf.Reset()
	}

	return nil
}

func findStartTag(cTag *Tag, tags []*Tag) *Tag {
	var pair int
	for i := len(tags) - 1; i >= 0; i-- {
		t := tags[i]
		if t.Name != cTag.Name || t.IsSelfClosing {
			continue
		}
		if t.IsEnd {
			pair++
			continue
		}
		if t.IsEnd == false && pair > 0 {
			pair--
			continue
		}

		if pair < 0 {
			return nil
		}

		if t.Name == cTag.Name {
			return t
		}
	}

	return nil
}

func componentTemplates(folder, ext string, funcMap template.FuncMap, readFile readFileFunc) (*template.Template, error) {
	t := template.New("").Funcs(funcMap)

	var filenames []string
	err := filepath.WalkDir(folder, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}

		if filepath.Ext(path) != ext {
			return nil
		}

		filenames = append(filenames, path)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return parseFiles(t, readFile, funcMap, filenames)
}
