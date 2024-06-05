package templates

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

func (t *Template) processComponentsInTemplate(contents *[]byte) error {
	// return nil

	if t.componentTemplates == nil {
		return nil
	}

	buf := bytes.NewBuffer(nil)
	tags := []*Tag{}
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

		// check if a template named t.name exists in the components folder
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
