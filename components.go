package templates

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type ComponentRenderer func(wr io.Writer, block *Block, tpl *template.Template) error

func (t *Template) processComponentsInTemplate(contents *[]byte) error {
	// if t.componentTemplates == nil {
	// 	return nil
	// }

	buf := bytes.NewBuffer(nil)
	for {
		block, err := FindInnerBlock(*contents)
		if err != nil {
			return err
		}
		if block == nil {
			break
		}

		// renderer, found := t.components[cName]
		compID := fmt.Sprintf("%s%d", block.Name, len(t.compData)+1)
		t.compData[compID] = &ComponentData{
			Args: block.Args,
			Body: block.Body,
		}

		if err = t.renderComponent(buf, block.Name, compID, block.Body); err != nil {
			return err
		}

		// replace rendered component with tag block
		start := block.Positions[BlkHeadPos].Start
		end := block.Positions[BlkEndPos].Stop
		*contents = append((*contents)[:start], append(buf.Bytes(), (*contents)[end:]...)...)
		// tHalf := (*contents)[:start]
		// bHalf := (*contents)[end:]
		// *contents = append(tHalf, append(buf.Bytes(), bHalf...)...)

		buf.Reset()
	}

	return nil
}

func (t *Template) renderComponent(buf *bytes.Buffer, name, componentID string, componentBody string) error {
	fle, err := os.Open(filepath.Join(t.componentsFolder, strings.ToLower(name)+t.ext))
	if err != nil {
		return fmt.Errorf("cant open component template: %w", err)
	}

	_, err = buf.WriteString(fmt.Sprintf("{{ $args%s := (componentData $ %q) }}\n", componentID, componentID))
	if err != nil {
		return fmt.Errorf("renderComponent %q: %w", name, err)
	}
	src, err := io.ReadAll(fle)
	if err != nil {
		return fmt.Errorf("renderComponent %q: %w", name, err)
	}

	src = bytes.Replace(src, []byte("@args."), []byte(fmt.Sprintf("$args%s.Args.", componentID)), -1)
	src = bytes.Replace(src, []byte("@body"), []byte(componentBody), 1)

	_, err = buf.Write(src)
	if err != nil {
		return fmt.Errorf("renderComponent %q : %w", name, err)
	}

	return nil
}
