package templates

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindBlock(t *testing.T) {
	tests := []struct {
		name    string
		content string
		block   *Block
		wantErr error
	}{
		{
			name: "block with body",
			content: `
			<Block name="foo">Hey!</Block>
			`,
			block: &Block{
				Name: "Block", Args: ArgMap{"name": "foo"}, Body: "Hey!", SelfClosing: false,
				Positions: []BlockPosition{{4, 22}, {26, 34}},
			},
		},
		{
			name: "self closing block",
			content: `
			<Block name="foo"/>
			`,
			block: &Block{
				Name: "Block", Args: ArgMap{"name": "foo"}, SelfClosing: true,
				Positions: []BlockPosition{{4, 23}, {4, 23}},
			},
		},
		{
			name: "block with missing closing tag",
			content: `
			<Block name="foo">Hey!
			`,
			block:   nil,
			wantErr: fmt.Errorf("cant find closing tag for %q", "Block"),
		},
		{
			name: "block with miss-spelt closing tag",
			content: `
			<Block name="foo">Hey!</block>
			`,
			block:   nil,
			wantErr: fmt.Errorf("cant find closing tag for %q", "Block"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			blk, err := FindBlock([]byte(tc.content), nil)
			if tc.wantErr != nil && err.Error() != tc.wantErr.Error() {
				t.Error(err)
			}

			assert.Equal(t, tc.block, blk)
		})
	}
}

func TestFindInnerBlock(t *testing.T) {
	tests := []struct {
		name    string
		content string
		block   *Block
		wantErr error
	}{
		{
			name: "single block with body",
			content: `
			<Block name="foo">Hey!</Block>
			`,
			block: &Block{
				Name: "Block", Args: ArgMap{"name": "foo"}, Body: "Hey!", SelfClosing: false,
				Positions: []BlockPosition{{4, 22}, {26, 34}},
			},
		},
		{
			name: "self closing block, followed by a regular block",
			content: `
			<Block name="foo"/>
			<Block name="foo">Hey!</Block>
			`,
			block: &Block{
				Name: "Block", Args: ArgMap{"name": "foo"}, SelfClosing: true,
				Positions: []BlockPosition{{4, 23}, {4, 23}},
			},
		},
		{
			name: "nested block",
			content: `
			<Card name="fooCard">
				A card
				<Block name="foo">Hey!</Block>
			</Card>
			`,
			block: &Block{
				Name: "Block", Args: ArgMap{"name": "foo"}, Body: "Hey!", SelfClosing: false,
				Positions: []BlockPosition{{41, 59}, {63, 71}},
			},
		},
		{
			name: "nested block with same name sibling",
			content: `
			<Card name="fooCard">
				A card
				<Block name="foo">Hey!</Block>
				<Block name="fooBar">Hey fooBar!</Block>
			</Card>
			`,
			block: &Block{
				Name: "Block", Args: ArgMap{"name": "fooBar"}, Body: "Hey fooBar!", SelfClosing: false,
				Positions: []BlockPosition{{76, 97}, {108, 116}},
			},
		},
		{
			name: "nested block with same name child",
			content: `
			<Card name="fooCard">
				A card
				<Block name="foo">
					<Block name="fooBar">Hey fooBar!</Block>
				</Block>
			</Card>
			`,
			block: &Block{
				Name: "Block", Args: ArgMap{"name": "fooBar"}, Body: "Hey fooBar!", SelfClosing: false,
				Positions: []BlockPosition{{65, 86}, {97, 105}},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			blk, err := FindInnerBlock([]byte(tc.content))
			if tc.wantErr != nil && err.Error() != tc.wantErr.Error() {
				t.Error(err)
			}

			assert.Equal(t, tc.block, blk)
		})
	}
}
