package templates

import (
	"bytes"
	"strings"
)

func (t *Template) processComponentsInTemplate(contents *[]byte) error {
	if t.components == nil {
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
		cTpl := t.components.Lookup(strings.ToLower(cTag.Name + t.ext))
		if cTpl == nil {
			// no template for this tag
			continue
		}

		if err = cTpl.Execute(buf, cTag); err != nil {
			return err
		}

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
