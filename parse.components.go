package templates

import (
	"fmt"
	"regexp"
	"strings"
)

var reTagHead = regexp.MustCompile(`<([A-Z][a-z-]+) *([\w\W]*?) *>`)
var reTagEnd = regexp.MustCompile(`</([A-Z][a-z-]+)>`)
var reTagArg = regexp.MustCompile(`([\w\-]+) *= *["|']([\w\W\s]*?)["|']`)

type tag struct {
	loc         []int
	Name        string
	Args        map[string]string
	Body        string
	SelfClosing bool
}

func findNextTag(content []byte) (*tag, error) {
	t, err := findTagHead(content)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, nil
	}

	if t.SelfClosing {
		t.loc = []int{t.loc[0], t.loc[1]}
		return t, nil
	}

	var blockEnd int
	t.Body, blockEnd, err = getBody(t.Name, content[t.loc[1]:])
	if err != nil {
		return nil, err
	}
	t.loc = []int{t.loc[0], t.loc[1] + blockEnd}

	return t, nil
}

func findTagHead(content []byte) (*tag, error) {
	loc := reTagHead.FindSubmatchIndex(content)
	if loc == nil {
		return nil, nil
	}

	t := new(tag)
	t.loc = loc
	t.Name = string(content[loc[2]:loc[3]])
	if len(loc) > 4 {
		args := string(content[loc[4]:loc[5]])
		t.Args = parseArgs(args)
	}

	head := string(content[loc[0]:loc[1]])
	if strings.HasSuffix(head, "/>") {
		t.SelfClosing = true
	}

	return t, nil
}

func parseArgs(args string) map[string]string {
	retv := map[string]string{}
	locs := reTagArg.FindAllIndex([]byte(args), -1)
	for _, l := range locs {
		p := strings.Split(args[l[0]:l[1]], "=")
		if len(p) != 2 {
			continue
		}

		val := strings.TrimRight(strings.TrimLeft(strings.TrimSpace(p[1]), `"''`), `"''`)
		retv[strings.TrimSpace(p[0])] = val
	}

	return retv
}

func getBody(tagName string, content []byte) (string, int, error) {
	locs := findTagHeads(tagName, content)
	headTagsFound := len(locs)

	// find closing tags
	locs = findTagEnds(tagName, content)
	if locs == nil || len(locs) != headTagsFound+1 {
		return "", -1, fmt.Errorf("cant find closing tag for <%s>", tagName)
	}

	// extract body
	lastEndTag := locs[len(locs)-1]
	body := string(content[:lastEndTag[0]])

	return body, lastEndTag[1], nil
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

func findTagHeads(tagName string, content []byte) [][]int {
	tagName = strings.ToLower(tagName)
	locs := reTagHead.FindAllIndex(content, -1)
	if locs == nil {
		return nil
	}

	var retv [][]int
	for _, l := range locs {
		ti := getTagInfo(reTagHead, l, content)
		if ti != nil && ti.name == tagName && !ti.selfClosing {
			retv = append(retv, l)
		}
	}

	return retv
}

func findTagEnds(tagName string, content []byte) [][]int {
	tagName = strings.ToLower(tagName)
	locs := reTagEnd.FindAllIndex(content, -1)
	if locs == nil {
		return nil
	}

	var retv [][]int
	for _, l := range locs {
		ti := getTagInfo(reTagEnd, l, content)
		if ti != nil && ti.name == tagName {
			retv = append(retv, l)
		}
	}

	return retv
}
