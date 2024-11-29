package templates

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/mayowa/templates/scanner"
)

var reTag = regexp.MustCompile(`</* *([A-Z][a-z-]+[a-zA-Z]*) *([\w\W]*?) *>`)

type ArgMap map[string]string

func (m ArgMap) ArgPairs() string {
	retv := []string{}
	for k, v := range m {
		v = strings.TrimSpace(v)
		if strings.HasPrefix(v, "{{") && strings.HasSuffix(v, "}}") {
			v = strings.Trim(v, "{}")
			retv = append(retv, fmt.Sprintf("%q %s", k, v))
		} else {
			retv = append(retv, fmt.Sprintf("%q %q", k, v))
		}
	}
	return strings.Join(retv, " ")
}

type Tag struct {
	loc           []int
	Name          string
	Args          ArgMap
	IsSelfClosing bool
	IsEnd         bool
}

func findNextTag(content []byte) (*Tag, error) {
	var err error
	loc := reTag.FindSubmatchIndex(content)
	if loc == nil {
		return nil, nil
	}

	t := new(Tag)
	t.loc = loc
	t.Name = string(content[loc[2]:loc[3]])
	t.IsSelfClosing = strings.HasSuffix(string(content[loc[0]:loc[1]]), "/>")
	t.IsEnd = strings.HasPrefix(string(content[loc[0]:loc[1]]), "</")

	// parse arguments
	if len(loc) > 4 && !t.IsEnd {
		args := string(content[loc[4]:loc[5]])

		scan := scanner.NewScanner(bytes.NewBufferString(args))
		t.Args, err = scan.ParseTagArgs()
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

func findAllTags(content []byte) (tags []*Tag, err error) {
	locs := reTag.FindAllSubmatchIndex(content, -1)

	for _, loc := range locs {
		t := new(Tag)
		t.loc = loc
		t.Name = string(content[loc[2]:loc[3]])
		t.IsSelfClosing = strings.HasSuffix(string(content[loc[0]:loc[1]]), "/>")
		t.IsEnd = strings.HasPrefix(string(content[loc[0]:loc[1]]), "</")

		// parse arguments
		if len(loc) > 4 && !t.IsEnd {
			args := string(content[loc[4]:loc[5]])

			scan := scanner.NewScanner(bytes.NewBufferString(args))
			t.Args, err = scan.ParseTagArgs()
			if err != nil {
				return nil, err
			}
		}
		tags = append(tags, t)
	}

	return tags, nil
}
