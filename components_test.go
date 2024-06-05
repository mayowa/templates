package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_findStartTag(t *testing.T) {

	tests := []struct {
		name string
		cTag *Tag
		tags []*Tag
		want *Tag
	}{
		{
			name: "best case",
			cTag: &Tag{Name: "foo", IsEnd: true},
			tags: []*Tag{
				{Name: "foo", IsEnd: false},
			},
			want: &Tag{Name: "foo", IsEnd: false},
		},
		{
			name: "nested other tag",
			cTag: &Tag{Name: "foo", IsEnd: true},
			tags: []*Tag{
				{Name: "foo", IsEnd: false},
				{Name: "bar", IsEnd: false},
				{Name: "bar", IsEnd: true},
			},
			want: &Tag{Name: "foo", IsEnd: false},
		},
		{
			name: "nested same tag",
			cTag: &Tag{Name: "foo", IsEnd: true},
			tags: []*Tag{
				{Name: "foo", IsEnd: false, loc: []int{0, 1}},
				{Name: "bar", IsEnd: false},
				{Name: "foo", IsEnd: false},
				{Name: "lala", IsSelfClosing: true},
				{Name: "foo", IsEnd: true},
				{Name: "bar", IsEnd: true},
			},
			want: &Tag{Name: "foo", IsEnd: false, loc: []int{0, 1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, findStartTag(tt.cTag, tt.tags), "findStartTag(%v, %v)", tt.cTag, tt.tags)
		})
	}
}
