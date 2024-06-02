package templates

import "testing"

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
