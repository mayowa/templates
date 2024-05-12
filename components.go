package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"strings"
)

type ComponentRenderer func(wr io.Writer, tag *Tag, tpl *template.Template) error

func (t *Template) processComponentsInTemplate(contents *[]byte) error {
	if t.componentTemplates == nil {
		return nil
	}

	buf := bytes.NewBuffer(nil)
	for {
		cTag, err := findNextTag(*contents)
		if err != nil {
			return err
		}
		if cTag == nil {
			break
		}

		// check if a template named t.name exists in the components folder
		cName := strings.ToLower(cTag.Name)

		args := fmt.Sprintf(`(map "_isSelfClosing" %v "_isEnd" %v %s)`, cTag.IsSelfClosing, cTag.IsEnd, cTag.Args.ArgPairs())
		buf.WriteString(`{{ template components/"` + cName + `" ` + args + ` }}`)

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
