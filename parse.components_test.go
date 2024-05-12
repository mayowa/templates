package templates

import (
	"testing"

	"github.com/go-test/deep"
)

func Test_findAllTags(t *testing.T) {
	tests := []struct {
		Name    string
		Content []byte
		Tags    []*Tag
		Error   error
	}{
		{
			Name: "card",
			Content: []byte(`
			< Card arg="arg1" age="22" >
			</Card >
			`),
			Tags: []*Tag{
				{
					Name: "Card",
					Args: map[string]string{"arg": "arg1", "age": "22"},
				},
				{
					Name:  "Card",
					IsEnd: true,
				},
			},
		},
		{
			Name: "deck",
			Content: []byte(`
			<Deck arg="arg1" />
			`),
			Tags: []*Tag{
				{
					Name:          "Deck",
					Args:          map[string]string{"arg": "arg1"},
					IsSelfClosing: true,
				},
			},
		},
		{
			Name: "card-deck",
			Content: []byte(`
			< Card arg="arg1" age="22" >
				<Deck arg="arg2" ></Deck >
			</Card>			
			<Deck arg="arg3" ></Deck>
			`),
			Tags: []*Tag{
				{
					Name: "Card",
					Args: map[string]string{"arg": "arg1", "age": "22"},
				},
				{
					Name: "Deck",
					Args: map[string]string{"arg": "arg2"},
				},
				{
					Name:  "Deck",
					IsEnd: true,
				},
				{
					Name:  "Card",
					IsEnd: true,
				},
				{
					Name: "Deck",
					Args: map[string]string{"arg": "arg3"},
				},
				{
					Name:  "Deck",
					IsEnd: true,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			tag, err := findAllTags(tt.Content)
			if err != tt.Error {
				t.Errorf("findNextTag() error = %v, wantErr %v", err, tt.Error)
			}

			if diff := deep.Equal(tag, tt.Tags); diff != nil {
				t.Error(diff)
			}
		})
	}
}
