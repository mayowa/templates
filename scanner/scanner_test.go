package scanner

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestScanner_Scan(t *testing.T) {

	tests := []struct {
		name  string
		input string
		want  *TokenItem
	}{
		{
			name:  "whitespace",
			input: "    Weee[]",
			want: &TokenItem{
				Token:         TokenWhiteSpace,
				Literal:       "    ",
				StartPosition: 0,
				EndPosition:   4,
				Line:          0,
			},
		},
		{
			name:  "identifier",
			input: "We-ee[]",
			want: &TokenItem{
				Token:         TokenIdentifier,
				Literal:       "We-ee",
				StartPosition: 0,
				EndPosition:   5,
				Line:          0,
			},
		},
		{
			name:  "tagStart",
			input: "<We-ee",
			want: &TokenItem{
				Token:         TokenTagStart,
				Literal:       "<",
				StartPosition: 0,
				EndPosition:   1,
				Line:          0,
			},
		},
		{
			name:  "closingTagStart",
			input: "</We-ee",
			want: &TokenItem{
				Token:         TokenClosingTagStart,
				Literal:       "</",
				StartPosition: 0,
				EndPosition:   2,
				Line:          0,
			},
		},
		{
			name:  "TagSelfClosing",
			input: "/>",
			want: &TokenItem{
				Token:         TokenTagSelfClosing,
				Literal:       "/>",
				StartPosition: 0,
				EndPosition:   2,
				Line:          0,
			},
		},
		{
			name:  "LeftAngleBracket",
			input: "<we-ee",
			want: &TokenItem{
				Token:         TokenLeftAngleBracket,
				Literal:       "<",
				StartPosition: 0,
				EndPosition:   1,
				Line:          0,
			},
		},
		{
			name:  "BackSlash",
			input: "/ ",
			want: &TokenItem{
				Token:         TokenBackSlash,
				Literal:       "/",
				StartPosition: 0,
				EndPosition:   1,
				Line:          0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(bytes.NewBufferString(tt.input))
			if got := s.Scan(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scan() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanner_ScanUntil(t *testing.T) {

	tests := []struct {
		name         string
		input        string
		token        Token
		withPeek     bool
		wantItems    []*TokenItem
		wantLastItem *TokenItem
	}{
		{
			name:  "test 1",
			input: `<Card a=1 />`,
			token: TokenTagSelfClosing,
			wantItems: []*TokenItem{
				{
					Token:         TokenTagStart,
					Literal:       "<",
					Line:          0,
					StartPosition: 0,
					EndPosition:   1,
				},
				{
					Token:         TokenIdentifier,
					Literal:       "Card",
					Line:          0,
					StartPosition: 1,
					EndPosition:   5,
				},
				{
					Token:         TokenWhiteSpace,
					Literal:       " ",
					Line:          0,
					StartPosition: 5,
					EndPosition:   6,
				},
				{
					Token:         TokenIdentifier,
					Literal:       "a",
					Line:          0,
					StartPosition: 6,
					EndPosition:   7,
				},
				{
					Token:         TokenAssign,
					Literal:       "=",
					Line:          0,
					StartPosition: 7,
					EndPosition:   8,
				},
				{
					Token:         TokenOther,
					Literal:       "1",
					Line:          0,
					StartPosition: 8,
					EndPosition:   9,
				},
				{
					Token:         TokenWhiteSpace,
					Literal:       " ",
					Line:          0,
					StartPosition: 9,
					EndPosition:   10,
				},
				{
					Token:         TokenTagSelfClosing,
					Literal:       "/>",
					Line:          0,
					StartPosition: 10,
					EndPosition:   12,
				},
			},
			wantLastItem: &TokenItem{
				Token:         TokenTagSelfClosing,
				Literal:       "/>",
				Line:          0,
				StartPosition: 10,
				EndPosition:   12,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(bytes.NewBufferString(tt.input))

			gotItems, gotLastItem := s.ScanUntil(tt.token, tt.withPeek)
			if !reflect.DeepEqual(gotItems, tt.wantItems) {
				t.Errorf("ScanUntil() gotItems = %v, want %v", dumpTokenItems(gotItems), dumpTokenItems(tt.wantItems))
			}
			if !reflect.DeepEqual(gotLastItem, tt.wantLastItem) {
				t.Errorf("ScanUntil() gotLastItem = %v, want %v", dumpTokenItem(gotLastItem), dumpTokenItem(tt.wantLastItem))
			}
		})
	}
}

func dumpTokenItem(item *TokenItem) string {
	if item == nil {
		return "Token: nil"
	}
	out := fmt.Sprintf("Token:%s\nLiteral:%s\nLine:%d\nStartPosition:%d\nEndPosition:%d\n", item.Token, item.Literal, item.Line, item.StartPosition, item.EndPosition)

	return out
}
func dumpTokenItems(items []*TokenItem) string {
	var out []string
	for _, i := range items {
		out = append(out, dumpTokenItem(i))
	}

	return strings.Join(out, "")
}
