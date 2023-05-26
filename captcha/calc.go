package captcha

import (
	"strconv"
	"strings"
	"unicode"
)

type Empty = struct{}

var ops = map[rune]Empty{
	'+': Empty{},
	'-': Empty{},
	'*': Empty{},
	'/': Empty{},
}

const EMPTY = 1 << 10

type Term struct {
	Left, Mid, Right float32
	F                rune
}

func NewTerm() *Term {
	return &Term{
		Left:  EMPTY,
		Mid:   EMPTY,
		Right: EMPTY,
		F:     '+',
	}
}

func (term *Term) Parse(str string) *Term {
	res := []rune(str)
	for i, r := range res {
		if !unicode.IsDigit(r) {
			_, ok := ops[r]
			if ok {
				term.F = r
			}
			if r != '?' {
				res[i] = rune(' ')
			}
		}
	}
	terms := strings.Fields(string(res))
	if len(terms) < 3 {
		return term
	}
	choose := func(word string) float32 {
		if word == "?" {
			return EMPTY
		}
		r, err := strconv.Atoi(word)
		if err != nil {
			return EMPTY
		}
		return float32(r)
	}
	term.Left = choose(terms[0])
	term.Mid = choose(terms[1])
	term.Right = choose(terms[2])
	return term
}

func (term *Term) Eval() float32 {
	switch term.F {
	case '+':
		if term.Right == EMPTY {
			return term.Left + term.Mid
		}
		if term.Left == EMPTY {
			return term.Right - term.Mid
		}
		if term.Mid == EMPTY {
			return term.Right - term.Left
		}
	case '-':
		term.F = '+'
		term.Mid = -term.Mid
		return term.Eval()
	case '*':
		if term.Right == EMPTY {
			return term.Left * term.Mid
		}
		if term.Left == EMPTY && term.Mid != 0 {
			return term.Right / term.Mid
		}
		if term.Mid == EMPTY && term.Left != 0 {
			return term.Right / term.Left
		}
	case '/':
		term.F = '*'
		if term.Mid != 0 && term.Mid != EMPTY {
			term.Mid = 1 / term.Mid
		} else if term.Mid == EMPTY && term.Right != 0 {
			return term.Left / term.Right
		}
		return term.Eval()
	}
	return 0

}

func Solve(str string) float32 {
	return NewTerm().Parse(str).Eval()
}
