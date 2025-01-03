package templates

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

func BuiltinInit(t *Template, options *TemplateOptions) error {
	t.FuncMap = options.FuncMap
	if t.FuncMap == nil {
		t.FuncMap = make(template.FuncMap)
	}

	// register the svg helper when path to svg is provided
	if options.PathToSVG != "" {
		t.FuncMap["svg"] = SvgHelper(options.PathToSVG)
	}

	t.FuncMap["html"] = func(v string) template.HTML { return template.HTML(v) }
	t.FuncMap["map"] = aMap
	t.FuncMap["slice"] = makeSlice
	t.FuncMap["component"] = t.component
	t.FuncMap["replaceStr"] = replaceStr
	t.FuncMap["ifZero"] = ifZero
	t.FuncMap["attributeSet"] = attributes
	t.FuncMap["deDupeStr"] = deDupeString
	t.FuncMap["mergeTwClasses"] = MergeTwClasses
	t.FuncMap["toJson"] = ToJson
	t.FuncMap["formatWithCommas"] = FormatWithCommas
	t.FuncMap["strInList"] = inList
	t.FuncMap["stringSet"] = stringSet

	return nil
}

func (t *Template) component(name string, args map[any]any) template.HTML {
	name += t.ext
	tpl := t.componentTemplates.Lookup(name)
	if tpl == nil {
		return ""
	}

	buff := bytes.NewBufferString("")
	if err := tpl.Execute(buff, args); err != nil {
		return template.HTML(err.Error())
	}

	return template.HTML(buff.String())
}

func aMap(args ...any) map[any]interface{} {
	retv := make(map[any]interface{}, len(args))
	for i := 0; i < len(args); i += 2 {
		retv[args[i]] = args[i+1]
	}

	return retv
}

func makeSlice(args ...any) []any {
	retv := make([]any, len(args))

	copy(retv, args)

	return retv
}

// returns a default value (second param)
// if first param is nil or a zero value
// else returns first param
func ifZero(src any, def any) any {
	if src == nil {
		return def
	}

	vs := reflect.ValueOf(src)
	vs = reflect.Indirect(vs)

	if vs.IsZero() {
		return def
	}

	return src
}

func replaceStr(str, old, new string) string {
	if old == "" {
		return str
	}

	return strings.Replace(str, old, new, -1)
}

type HTMLAttributes map[string]string

func attributes() *HTMLAttributes {
	a := make(HTMLAttributes)
	return &a
}

func (a *HTMLAttributes) Set(key, value string) string {
	(*a)[key] = value
	return ""
}

func (a *HTMLAttributes) Render() template.HTMLAttr {
	out := ""
	for k, v := range *a {
		out += fmt.Sprintf(" %s=%q ", k, v)
	}

	return template.HTMLAttr(out)
}

// deDupe removes duplicate substrings from the source string
// separator is space by default but can be specified otherwise
func deDupeString(src string, argv ...string) string {
	sep := " "
	if len(argv) > 0 && argv[0] != "" {
		sep = argv[0]
	}

	slice := strings.Split(src, sep)
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range slice {
		if _, ok := seen[item]; !ok {
			seen[item] = true
			result = append(result, item)
		}
	}

	return strings.Join(result, sep)
}

var reClassAttr = regexp.MustCompile(`(?m) class *= *["']([^"']*)["']`)
var reAttr = regexp.MustCompile(`([\w-]*) *= *["'][\w -]+["']`)

// SvgHelper expects that the svg markup in the specified file has a class attribute, even if it's empty
func SvgHelper(folder string) func(name string, class ...string) template.HTML {
	return func(name string, class ...string) template.HTML {

		cls := ""
		if len(class) > 0 {
			cls = class[0]
		}
		file := filepath.Join(folder, name+".svg")
		contents, err := os.ReadFile(file)
		if err != nil {
			return ""
		}

		// append supplied class to class on svg
		if cls != "" {
			var classStr string
			matches := reClassAttr.FindStringSubmatch(string(contents))

			if len(matches) >= 2 {
				classStr = fmt.Sprintf(" class=%q", matches[1]+" "+cls)
				// contents = reClassAttr.ReplaceAll(contents, []byte(classStr))
				contents = bytes.Replace(contents, []byte(matches[0]), []byte(classStr), 1)
			}
		}

		// remove width and height attributes
		// limit search to the head of the svg
		attrs := findAttributes(string(bytes.Split(contents, []byte(">"))[0]))
		for _, a := range attrs {
			if a[0] == "width" || a[0] == "height" {
				contents = bytes.Replace(contents, []byte(a[1]), []byte(""), 1)
			}
		}

		return template.HTML(contents)
	}
}

func findAttributes(src string) [][2]string {
	var pairs [][2]string
	ret := reAttr.FindAllStringSubmatchIndex(src, -1)
	for _, v := range ret {
		// fmt.Println(i, src[v[0]:v[1]], src[v[2]:v[3]])
		pairs = append(pairs, [2]string{src[v[2]:v[3]], src[v[0]:v[1]]})
	}

	return pairs
}

// Merge Tailwind classes builtin
/**
Input: two strings s1 and s2. each string consists of substrings separated by spaces. each substring itself consists of substrings separated by hyphens
Output: single string representing a union of the input strings. this union is formed by adding the substrings according to the following rules:
- add each substring in s1 to the output. unless it conflicts with a substring in s2. each s1 substring will conflict with at most one substring in s2
- add each substring in s2 to the output, unless it conflicts with a substring in s1.
- when there's a conflict, discard the substring from s2

a conflict is defined as the situation where two substrings are either identical or identical except for their last component substring.
for example, px-5 and px-4 conflict, because they're identical except for their last component.
p-5 and px-5 do not conflict, b-t-1 and b-s-2 do not conflict, b-t-1 and b-1 do not conflict.
strings without a hyphen will always conflict.

**/

// Helper function to extract the base class without the last component
func getBaseClass(class string, sep string) string {
	parts := strings.Split(class, sep)
	if len(parts) == 1 {
		return class
	}
	return strings.Join(parts[:len(parts)-1], sep)
}

func MergeTwClasses(priority, def, sep string) string {
	if priority == "" {
		return def
	}

	if def == "" {
		return priority
	}

	if sep == "" {
		sep = " "
	}

	// Split the input strings into classes
	priorityClasses := strings.Split(priority, sep)
	defaultClasses := strings.Split(def, sep)

	// Maps to track classes and their base forms
	classMap := make(map[string]bool)
	baseMap := make(map[string]string)

	// Process priority classes
	for _, pClass := range priorityClasses {
		classMap[pClass] = true
		baseClass := getBaseClass(pClass, "-")
		baseMap[baseClass] = pClass
	}

	// Result list containing merged classes
	res := make([]string, 0, len(priorityClasses))
	res = append(res, priorityClasses...)

	// Process default classes and check for conflicts
	for _, dClass := range defaultClasses {
		baseClass := getBaseClass(dClass, "-")
		if _, exists := baseMap[baseClass]; exists {
			// Conflict: Skip this default class
			continue
		}
		if _, exists := classMap[dClass]; exists {
			// Exact match conflict: Skip this default class
			continue
		}

		// No conflict: Add to result and update maps
		res = append(res, dClass)
		classMap[dClass] = true
		baseMap[baseClass] = dClass
	}

	return strings.Join(res, sep)
}

// stringSet given a @src string containing substrings separated with (@sep)
// return a union of substrings in @src and @add
func stringSet(src, sep, add string) string {
	var srcList []string
	src = strings.TrimSpace(src)
	add = strings.TrimSpace(add)

	if src != "" {
		srcList = strings.Split(src, sep)
	}

	parts := strings.Split(add, sep)

	for _, str := range parts {
		if inList(str, srcList) {
			continue
		}

		srcList = append(srcList, str)
	}

	return strings.Join(srcList, sep)
}

func inList(cmp string, list []string) bool {
	cmp = strings.ToLower(cmp)
	for _, str := range list {
		if cmp == strings.ToLower(str) {
			return true
		}
	}
	return false
}

func FormatWithCommas(n any) string {
	var str string

	// Check if n is an int or float64 and format accordingly
	switch v := n.(type) {
	case int:
		str = strconv.Itoa(v)
	case float64:
		// Format the float with up to 2 decimal places
		str = fmt.Sprintf("%.2f", v)
	default:
		return "Invalid type"
	}

	// Split integer and fractional parts for floats
	parts := strings.Split(str, ".")
	integerPart := parts[0]
	length := len(integerPart)

	// Add commas to the integer part
	if length > 3 {
		var result strings.Builder
		firstGroupSize := length % 3
		if firstGroupSize == 0 {
			firstGroupSize = 3
		}
		result.WriteString(integerPart[:firstGroupSize])
		for i := firstGroupSize; i < length; i += 3 {
			result.WriteString(",")
			result.WriteString(integerPart[i : i+3])
		}
		integerPart = result.String()
	}

	// If it's a float, reattach the fractional part
	if len(parts) > 1 {
		return integerPart + "." + parts[1]
	}

	return integerPart
}

func ToJson(v any) string {
	val, err := json.Marshal(v)
	if err != nil {
		return err.Error()
	}

	return string(val)
}
