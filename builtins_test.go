package templates

import "testing"

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := deDupeString(tt.src, tt.sep); got != tt.expect {
				t.Errorf("deDupeString() = %v, want %v", got, tt.expect)
			}
		})
	}
}
