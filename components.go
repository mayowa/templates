package templates

import (
	"fmt"
	"regexp"
	"strings"
)

var reTagHead = regexp.MustCompile(`<([A-Z][a-z-]+) *([\w\W]*?) *>`)
var reTagEnd = regexp.MustCompile(`</([A-Z][a-z-]+)>`)

func findTagBlock(content []byte) (block []byte, err error) {
	// findHead := func(content []byte) []byte {
	// 	loc := reTagHead.FindIndex(content)
	// 	if loc == nil {
	// 		return nil
	// 	}
	// 	return content[loc[0]:loc[1]]
	// }
	// findEnd := func(content []byte) []byte {
	// 	loc := reTagEnd.FindIndex(content)
	// 	if loc == nil {
	// 		return nil
	// 	}
	// 	return content[loc[0]:loc[1]]
	// }
	//
	// reTagHead.FindSubmatchIndex()
	//
	//

	return
}

package main

import (
"fmt"
"regexp"
"strings"
)

var reTagHead = regexp.MustCompile(`<([A-Z][a-z-]+) *([\w\W]*?) *>`)
var reTagEnd = regexp.MustCompile(`</([A-Z][a-z-]+)>`)

var text = []byte(`
<Card foo="age" class="bg-red-200">
  <Card >Hey!</Card>
	<Lard>Say</Lard>
</Card>
`)

func main() {
	// loc := reTagHead.FindIndex(text)
	// fmt.Println(string(text[loc[0]:loc[1]]))
	//
	// loc = reTagHead.FindSubmatchIndex(text)
	// fmt.Println(loc)
	// fmt.Println(string(text[loc[0]:loc[1]]))
	// fmt.Println(string(text[loc[2]:loc[3]]))
	// fmt.Println(string(text[loc[4]:loc[5]]))
	//
	// locs := reTagHead.FindAllIndex(text, -1)
	// fmt.Println(len(locs))
	// for _, l := range locs {
	// 	fmt.Println(l)
	// }

	t, err := findNextTag(text)
	if err != nil {
		fmt.Println(err)
		return
	}
	if t == nil {
		fmt.Println("no tag found")
		return
	}
	// fmt.Println(*t)
	fmt.Println(t.name, ":", string(text[t.loc[0]:t.loc[1]]))
}

type tag struct {
	loc []int
	name string
	args map[string]string
	body string
	selfClosing bool
}

func findNextTag(content []byte) (*tag, error) {
	t, err := findTagHead(content)
	if err != nil {
		return nil, err
	}
	if t == nil || t.selfClosing{
		return nil, nil
	}

	var blockEnd int
	t.body, blockEnd, err = getBody(t.name, content[t.loc[1]:])
	if err != nil {
		return nil, err
	}
	t.loc = []int{t.loc[0], t.loc[1]+blockEnd}

	return t, nil
}

func findTagHead(content []byte) (*tag, error) {
	loc := reTagHead.FindSubmatchIndex(content)
	if loc == nil {
		return nil, nil
	}

	t := new(tag)
	t.loc = loc
	t.name = string(content[loc[2]:loc[3]])
	if len(loc) > 4 {
		args := string(content[loc[4]:loc[5]])
		t.args = parseArgs(args)
	}

	head := string(text[loc[0]:loc[1]])
	if strings.HasSuffix(head, "/>") {
		t.selfClosing = true
	}

	return t, nil
}

func parseArgs(args string) map[string]string {
	pArgs := strings.Fields(args)

	retv := map[string]string{}
	for _, arg := range pArgs {
		p := strings.Split(arg, "=")
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

func findTagEndLocation(content []byte) []int {
	loc := reTagEnd.FindIndex(content)
	return loc
}

type tagInfo struct {
	name string
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
		if  ti != nil && ti.name == tagName && !ti.selfClosing {
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