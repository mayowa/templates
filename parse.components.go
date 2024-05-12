package templates

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/mayowa/templates/scanner"
)

var reTag = regexp.MustCompile(`</* *([A-Z][a-z-]+) *([\w\W]*?) *>`)

type Tag struct {
	loc           []int
	Name          string
	Args          map[string]string
	IsSelfClosing bool
	IsEnd         bool
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

type tagInfo struct {
	name        string
	selfClosing bool
}

func getTagInfo(re *regexp.Regexp, l []int, txt []byte) *tagInfo {
	nTxt := txt[l[0]:l[1]]
	loc := re.FindSubmatchIndex(nTxt)
	if loc == nil {
		return nil
	}
	ti := &tagInfo{
		name:        strings.ToLower(string(nTxt[loc[2]:loc[3]])),
		selfClosing: strings.HasSuffix(string(nTxt[loc[0]:loc[1]]), "/>"),
	}

	return ti
}
