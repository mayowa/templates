package scanner

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func TestScanner_ParseTagArgs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantArgs map[string]string
		wantErr  error
	}{
		{
			name:     "with nested double quotes",
			input:    `name=""ayo" and "chidy"" age="21" gen-der="m"`,
			wantArgs: map[string]string{"name": `"ayo" and "chidy"`, "age": "21", "gen-der": "m"},
		},
		{
			name:     "with template tags",
			input:    `name="{{if eq a "ayo"}} and "chidy"{{end}}" age="21" gen-der="m"`,
			wantArgs: map[string]string{"name": `{{if eq a "ayo"}} and "chidy"{{end}}`, "age": "21", "gen-der": "m"},
		},
		{
			name:    "with unterminated string",
			input:   `name="ayo `,
			wantErr: errors.New("extractArgVal() found unterminated string on line:0"),
		},
		{
			name:    "with double unterminated string",
			input:   `name=""ayo"`,
			wantErr: errors.New("extractArgVal() found unterminated string on line:0"),
		},
		{
			name:    "with missing quotes",
			input:   `name="ayo" and age="21" gen-der="m"`,
			wantErr: errors.New(`expected '=', found "age"`),
		},
		{
			name:     "with empty string",
			input:    `name=""`,
			wantArgs: map[string]string{"name": ""},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewScanner(bytes.NewBufferString(tt.input))

			gotArgs, gotErr := s.ParseTagArgs()
			if tt.wantErr != nil && gotErr.Error() != tt.wantErr.Error() {
				t.Fatalf("ParseTagArgs() gotErr = %v, wantErr %v", gotErr, tt.wantErr)
			} else if tt.wantErr == nil && gotErr != nil {
				t.Fatalf("Unexpected error in ParseTagArgs(), gotErr = %v", gotErr)
			}

			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("ParseTagArgs() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}

		})
	}
}
