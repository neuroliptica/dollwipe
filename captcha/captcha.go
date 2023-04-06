// captcha.go: main captcha datatypes and declarations.
// Solvers implementations can be found in eponymous files.

package captcha

import (
	"fmt"
)

// Error codes.
const (
	CAPTCHA_FAIL = iota
	CAPTCHA_NEED_CHECK
	CAPTCHA_PASSCODE
	CAPTCHA_PASSCODE_EXP
	CAPTCHA_HTTP_FAIL
)

// General solver function type.
// Every captcha's solver function should satisfy this signature.
type Solver = func(img []byte, key string) (string, error)

// 2ch API captcha id response.
type CaptchaJSON struct {
	Id, Input, Type string
	Result          int32
}

// 2ch.hk/abu/res/42375.html: catching captcha id response errors.
type CaptchaIdError struct {
	ErrorId int32
	Extra   error
}

// New CaptchaIdError instance with code from the response.
func NewCaptchaIdError(errorId int32, err error) *CaptchaIdError {
	return &CaptchaIdError{errorId, err}
}

// Error interface instance.
func (e *CaptchaIdError) Error() string {
	return fmt.Sprintf("%d", e.ErrorId)
}
