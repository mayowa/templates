package templates

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/mayowa/templates/scanner"
)

var reTagHead = regexp.MustCompile(`<([A-Z][a-z-]+) *([\w\W]*?) *>`)
var reTagEnd = regexp.MustCompile(`</([A-Z][a-z-]+)>`)
var reTagArg = regexp.MustCompile(`([\w\-]+) *= *["|']([\w\W\s]*?)["|']`)
var reGoTplTag = regexp.MustCompile(`{{[\s\S\w]*?}}`)

type ArgMap map[string]string

func (m ArgMap) ArgPairs() string {
	retv := []string{}
	for k, v := range m {
		retv = append(retv, fmt.Sprintf("%q %q", k, v))
	}
	return strings.Join(retv, " ")
}

type Tag struct {
	loc         []int
	Name        string
	Args        ArgMap
	Body        string
	SelfClosing bool
}

func findNextTag(content []byte) (*Tag, error) {
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

func NewTag(content []byte, loc []int) (tag *Tag, err error) {
	tag = new(Tag)
	tag.loc = loc
	tag.Name = string(content[loc[2]:loc[3]])
	tag.SelfClosing = strings.HasSuffix(string(content[loc[0]:loc[1]]), "/>")
	if len(loc) > 4 {
		args := string(content[loc[4]:loc[5]])
		if tag.SelfClosing {
			args = string(content[loc[4] : loc[5]-1])
		}
		scan := scanner.NewScanner(bytes.NewBufferString(args))
		tag.Args, err = scan.ParseTagArgs()
		if err != nil {
			return nil, err
		}
	}

	head := string(content[loc[0]:loc[1]])
	if strings.HasSuffix(head, "/>") {
		tag.SelfClosing = true
	}

	return tag, nil
}

func findTagHead(content []byte) (*Tag, error) {
	var (
		err error
		tag *Tag
	)

	loc := reTagHead.FindSubmatchIndex(content)
	if loc == nil {
		return nil, nil
	}

	if tag, err = NewTag(content, loc); err != nil {
		return nil, err
	}

	return tag, nil
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
	loc         []int
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
		loc:         l,
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

func findAllTagHeads(content []byte) []*tagInfo {
	locs := reTagHead.FindAllSubmatchIndex(content, -1)
	if locs == nil {
		return nil
	}

	var retv []*tagInfo
	for _, l := range locs {
		ti := getTagInfo(reTagHead, l, content)
		if ti != nil && !ti.selfClosing {
			retv = append(retv, ti)
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

func findAllTagEnds(content []byte, tagName string) []*tagInfo {
	tagName = strings.ToLower(tagName)
	locs := reTagEnd.FindAllIndex(content, -1)
	if locs == nil {
		return nil
	}

	var retv []*tagInfo
	for _, l := range locs {
		ti := getTagInfo(reTagEnd, l, content)
		if ti != nil {
			retv = append(retv, ti)
			if ti.name == tagName {
				break
			}
		}
	}

	return retv
}

type Block struct {
	Positions   []BlockPosition
	Name        string
	Args        ArgMap
	Body        string
	SelfClosing bool
}

type BlockPositions int

const (
	BlkHeadPos BlockPositions = iota
	BlkEndPos
)

type BlockPosition struct {
	Start int
	Stop  int
}

func FindInnerBlock(content []byte) (*Block, error) {
	var (
		err error
		tag *Tag
	)

	tag, err = findTagHead(content)
	if err != nil {
		return nil, err
	}
	if tag == nil {
		return nil, nil
	}
	if tag.SelfClosing {
		return FindBlock(content, tag)
	}

	// find tagHeads
	heads := findAllTagHeads(content)
	if heads == nil {
		return nil, nil
	}

	// proc innermost tag
	innerHead := heads[len(heads)-1]

	// find tagEnds for the innerHead most tag
	ends := findAllTagEnds(content, innerHead.name)
	if ends == nil {
		return nil, fmt.Errorf("can't find closing tag for <%s>", innerHead.name)
	}
	innerEnd := ends[0]

	if innerHead.name != innerEnd.name {
		return nil, fmt.Errorf("expected </%s> found </%s>", innerHead.name, innerEnd.name)
	}

	tag, err = NewTag(content, innerHead.loc)
	if err != nil {
		return nil, err
	}

	block, err := FindBlock(content, tag)
	if err != nil {
		return nil, err
	}

	return block, err
}

func FindBlock(content []byte, tag *Tag) (*Block, error) {
	var err error
	block := &Block{Positions: make([]BlockPosition, 2)}
	if tag == nil {
		tag, err = findTagHead(content)
	}
	if err != nil {
		return nil, err
	}
	if tag == nil {
		return nil, nil
	}

	block.Args = tag.Args
	block.Name = tag.Name
	block.SelfClosing = tag.SelfClosing
	block.Positions[BlkHeadPos].Start = tag.loc[0]
	block.Positions[BlkHeadPos].Stop = tag.loc[1]

	if block.SelfClosing {
		block.Positions[BlkEndPos].Start = tag.loc[0]
		block.Positions[BlkEndPos].Stop = tag.loc[1]

		return block, nil
	}

	ti, loc := findTagEnd(content[block.Positions[BlkHeadPos].Stop:])
	if ti == nil || ti.name != strings.ToLower(block.Name) {
		return nil, fmt.Errorf("cant find closing tag for %q", block.Name)
	}

	block.Positions[BlkEndPos].Start = loc[0] + block.Positions[BlkHeadPos].Stop
	block.Positions[BlkEndPos].Stop = loc[1] + block.Positions[BlkHeadPos].Stop
	block.Body = string(content[block.Positions[BlkHeadPos].Stop:block.Positions[BlkEndPos].Start])

	return block, nil
}

func findTagEnd(content []byte) (*tagInfo, []int) {
	loc := reTagEnd.FindSubmatchIndex(content)
	if loc == nil {
		return nil, nil
	}

	return getTagInfo(reTagEnd, loc, content), loc
}
