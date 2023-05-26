package captcha

import "testing"

func TestSolveTableDriven(t *testing.T) {
	var tests = []struct {
		name  string
		input string
		want  float32
	}{
		{"1", "1+1=?", 2},
		{"2", "?+4=42", 38},
		{"3", "10/?=2", 5},
		{"4", "7*?=21", 3},
		{"5", "7*3=?", 21},
		{"6", "1*9=?", 9},
		{"7", "3*?=24", 8},
		{"8", "7*?=21", 3},
		{"9", "62-28=?", 34},
		{"10", "?*1=4", 4},
		{"11", "?/1=20", 20},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans := Solve(tt.input)
			if ans != tt.want {
				t.Errorf("%f != %f", ans, tt.want)
			}
		})
	}
}
