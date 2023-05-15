package captcha

import "testing"

func TestReplaceTableDriven(t *testing.T) {
	//MustInitDict()
	var tests = []struct {
		name  string
		input string
		want  string
	}{
		{"1", "ф+зик", "ф?зик"},
		{"2", "ко+ьк+и", "ко?ьки"},
		{"3", "а+шла++г", "а?шлаг"},
		{"4", "жл++ый", "жл?ый"},
		{"5", "дру??ба", "дру?ба"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans := Replace(tt.input)
			if ans != tt.want {
				t.Errorf("%s != %s", ans, tt.want)
			}
		})
	}
}

func TestMatchTableDriven(t *testing.T) {
	MustInitDict("../res/dict")
	var tests = []struct {
		name  string
		input string
		want  string
	}{
		{"1", "ф+зик", "физик"},
		{"2", "руга?ь", "ругань"},
		{"3", "а+шла++г", "аншлаг"},
		{"4", "вые?д", "выезд"},
		{"5", "дру??ба", "дружба"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans := Match(tt.input)
			if ans != tt.want {
				t.Errorf("%s != %s", ans, tt.want)
			}
		})
	}
}
