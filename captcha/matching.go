// matching.go: matching result captcha values in dict.

package captcha

import "strings"

// Convert to format.
func Replace(value string) string {
	r := []rune(value)
	has := false
	p := func(d rune) bool { return d == rune('?') || d == rune('+') }
	for i := range r {
		if p(r[i]) {
			if has {
				r[i] = rune(' ')
			} else {
				r[i] = rune('?')
				has = true
			}
		}
	}
	n := string(r)
	return strings.Join(strings.Split(n, " "), "")
}

// Will try to match captcha value in dictionary.
func Match(value string) string {
	value = Replace(value)
	a := "абвгдеёжзийклмнопрстуфхцчшщъыьэюя"
	runes := []rune(a)
	parts := strings.Split(value, "?")
	if len(parts) != 2 {
		return value
	}
	for _, r := range runes {
		left := []rune(parts[0])
		left = append(left, r)
		left = append(left, []rune(parts[1])...)
		word := string(left)
		if _, ok := Dict[word]; ok {
			return word
		}
	}
	return value
}
