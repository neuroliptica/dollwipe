package engine

import (
	"fmt"
	"sync"
)

type (
	Captcha struct {
		Hash   string
		Base64 string
		Value  string
	}

	UnsolvedType struct {
		Captchas []Captcha
		Hashes   map[string]struct{}
	}
)

var (
	Unsolved = UnsolvedType{
		Captchas: make([]Captcha, 0),
		Hashes:   make(map[string]struct{}),
	}
	Queued = make(map[string]struct{})
	Solved = make(map[string]string)

	DataMutex sync.Mutex
)

func (u *UnsolvedType) Add(c Captcha) {
	u.Captchas = append(u.Captchas, c)
	u.Hashes[c.Hash] = struct{}{}
}

func (u *UnsolvedType) Pop() (Captcha, error) {
	if len(u.Captchas) == 0 {
		return Captcha{}, fmt.Errorf("NO_CAPTCHAS_AVAIBLE")
	}
	captcha := u.Captchas[0]
	u.Captchas = u.Captchas[1:]
	delete(u.Hashes, captcha.Hash)

	return captcha, nil
}

func (u *UnsolvedType) Has(hash string) bool {
	_, ok := u.Hashes[hash]
	return ok
}
