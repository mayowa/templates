package templates

import (
	"fmt"
	"regexp"
	"strings"
	"testing"
)

func Test__ifZero(t *testing.T) {
	tests := []struct {
		name   string
		src    any
		def    any
		expect any
	}{
		{
			name:   "src is nil",
			src:    nil,
			def:    "default",
			expect: "default",
		},

		{
			name:   "src is empty string",
			src:    "",
			def:    "default",
			expect: "default",
		},

		{
			name:   "src is zero",
			src:    "",
			def:    "default",
			expect: "default",
		},

		{
			name:   "src is not empty string",
			src:    "src",
			def:    "default",
			expect: "src",
		},

		{
			name:   "default is not str",
			src:    "",
			def:    0,
			expect: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ifZero(tt.src, tt.def); got != tt.expect {
				t.Errorf("ifZero() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func Test__deDupeString(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		sep    string
		expect string
	}{
		{
			name:   "empty",
			expect: "",
		},
		{
			name:   "one duplicate with name partial",
			src:    "red bg-red-300 green red",
			expect: "red bg-red-300 green",
		},
		{
			name:   "doesn't modify src if no duplicates",
			src:    "red bg-red-300 green",
			expect: "red bg-red-300 green",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deDupeString(tt.src, tt.sep); got != tt.expect {
				t.Errorf("deDupeString() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func Test__replaceStr(t *testing.T) {
	tests := []struct {
		name   string
		src    string
		old    string
		new    string
		expect string
	}{
		{
			name:   "replaces old with new",
			src:    "foo bar baz foo",
			old:    "foo",
			new:    "boo",
			expect: "boo bar baz boo",
		},

		{
			name:   "removes old when new is empty",
			src:    "foo bar baz",
			old:    "foo",
			new:    "",
			expect: " bar baz",
		},

		{
			name:   "doesn't modify string when old is empty",
			src:    "foo bar baz",
			old:    "",
			new:    "boo",
			expect: "foo bar baz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := replaceStr(tt.src, tt.old, tt.new); got != tt.expect {
				t.Errorf("replaceStr() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func Test__attrRender(t *testing.T) {
	attrs := map[string]string{
		"foo": "bar",
		"abc": "def",
		"key": "value",
	}

	attrMap := attributes()

	for k, v := range attrs {
		attrMap.Set(k, v)
	}

	renderedAttrs := string(attrMap.Render())
	if renderedAttrs[len(renderedAttrs)-1] != ' ' {
		t.Errorf("Trailing whitespace not found in rendered output %s", renderedAttrs)
	}

	for k, v := range attrs {
		attrPair := fmt.Sprintf("%s=%q", k, v)
		if !strings.Contains(renderedAttrs, attrPair) {
			t.Errorf("Key-Value pair %s not found in rendered string %s", attrPair, renderedAttrs)
		}
	}
}

func Test__SvgHelper(t *testing.T) {
	formattingRegex := regexp.MustCompile(`[\n\t]+`)

	tests := []struct {
		name     string
		svg      string
		class    string
		expected string
	}{
		{
			name:     "renders accurately",
			svg:      "down-chevron",
			expected: `<svg class="w-4 h-4 ms-3" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 10 6"><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m1 1 4 4 4-4"/></svg>`,
		},

		{
			name:     "appends classes correctly",
			svg:      "down-chevron",
			class:    "dummy-class",
			expected: `<svg class="w-4 h-4 ms-3 dummy-class" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 10 6"><path stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m1 1 4 4 4-4"/></svg>`,
		},

		{
			name:     "only appends to first class",
			svg:      "chevron-extra-class",
			class:    "dummy-class",
			expected: `<svg class="w-4 h-4 ms-3 dummy-class" aria-hidden="true" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 10 6"><path stroke="currentColor" class="unchanged" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="m1 1 4 4 4-4"/></svg>`,
		},

		{
			name:     "removes width and height correctly",
			svg:      "attr",
			expected: `<svg   dummy-width="gotcha" class="w-4 h-4 ms-3"><path stroke="currentColor" d="m1 1 4 4 4-4"/></svg>`,
		},

		{
			name:     "width and height on inner elems are untouched",
			svg:      "calendar",
			expected: `<svg xmlns="http://www.w3.org/2000/svg"   viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-calendar"><path d="M8 2v4"/><path d="M16 2v4"/><rect width="18" height="18" x="3" y="4" rx="2"/><path d="M3 10h18"/></svg>`,
		},
	}

	svgFunc := SvgHelper("./testData/svg")
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svgOut := string(svgFunc(tt.svg, tt.class))
			svgOut = formattingRegex.ReplaceAllString(svgOut, "")

			expected := formattingRegex.ReplaceAllString(tt.expected, "")

			if svgOut != expected {
				t.Errorf("Expected \n %s\n, got \n %s", expected, svgOut)
			}
		})
	}
}

func Test__MergeTwClasses(t *testing.T) {
	tests := []struct {
		name     string
		priority string
		def      string
		expected string
	}{
		{
			name:     "s1 overrides conflicts",
			priority: "px-5",
			def:      "px-4",
			expected: "px-5",
		},
		{
			name:     "s1 and s2 merge if no conflicts",
			priority: "p-5",
			def:      "px-4",
			expected: "p-5 px-4",
		},
		{
			name:     "",
			priority: "b-t-1",
			def:      "b-s-2",
			expected: "b-t-1 b-s-2",
		},
		{
			name:     "",
			priority: "b-t-1",
			def:      "b-1",
			expected: "b-t-1 b-1",
		},
		{
			name:     "",
			priority: "px-5 mb-2",
			def:      "px-4 mt-2",
			expected: "px-5 mb-2 mt-2",
		},
		{
			name:     "",
			priority: "px-5 mb-2",
			def:      "px-4 mb-5 mt-3",
			expected: "px-5 mb-2 mt-3",
		},
		{
			name:     "",
			priority: "",
			def:      "default str",
			expected: "default str",
		},
		{
			name:     "",
			priority: "pri str",
			def:      "",
			expected: "pri str",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := MergeTwClasses(tt.priority, tt.def, " ")
			if res != tt.expected {
				t.Errorf("Expected \n %s\n, got \n %s", tt.expected, res)
			}
		})
	}
}

func Test__stringSet(t *testing.T) {
	tests := []struct {
		name string
		src  string
		sep  string
		add  string
		want string
	}{
		{
			name: "empty",
		},
		{
			name: "test 1",
			src:  "single word",
			sep:  " ",
			add:  "and single foo Word",
			want: "single word and foo",
		},
		{
			name: "test 2",
			src:  "",
			sep:  " ",
			add:  "and single foo Word",
			want: "and single foo Word",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := stringSet(tt.src, tt.sep, tt.add)
			if res != tt.want {
				t.Errorf("Expected: %q, got: %q", tt.want, res)
			}
		})
	}
}
