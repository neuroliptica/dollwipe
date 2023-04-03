// engine.go: main posting logic.
// Coherently appling Post's methods and catching errors.

package engine

import (
	"dollwipe/captcha"
	"dollwipe/env"
	"dollwipe/network"
	"fmt"
	"net/http"
	"os"
)

// Because we also need to store the proxy.
type InitPostResponse struct {
	PostPtr *Post
	Proxy   network.Proxy
}

func (r InitPostResponse) Address() string {
	return r.Proxy.Addr
}

func (r InitPostResponse) Post() *Post {
	return r.PostPtr
}

func InitPost(penv *env.Env, proxy network.Proxy, ch chan<- InitPostResponse) {
	response := InitPostResponse{nil, proxy}
	post := Post{
		Env:        penv,
		Proxy:      proxy,
		HTTPFailed: 0,
		Headers:    make(map[string]env.Header, 0),
		Cookies:    make([]*http.Cookie, 0),
	}
	post.Log("получаю печенюшки...")
	// Will create browser instance, should be parallel.
	// Also set up proxy for browser instance.
	post.GetHeaders()
	if len(post.Cookies) == 0 {
		ch <- response
		return
	}
	response.PostPtr = &post
	for key, value := range post.Headers {
		post.Verbose(key, ": ", string(value))
	}
	for i := range post.Cookies {
		post.Verbose("Cookie: ", post.Cookies[i].String())
	}
	ch <- response
}

// Handle makaba's posting response.
func responseHandler(post *Post, code int32) {
	if code == OK {
		post.Env.Status <- true
		return
	}
	switch code {
	case ERROR_BANNED:
		post.Log("прокся забанена, удаляю.")
		post.Env.Filter <- post.Proxy.Addr
	case ERROR_ACCESS_DENIED:
		post.Log("доступ заблокирован, удаляю проксю.")
		post.Env.Filter <- post.Proxy.Addr
	case ERROR_CLOSED:
		post.Log("тред закрыт, маладца.")
		if post.Env.WipeMode == env.SINGLE {
			post.Log("больше не могу постить в этот тред, пора на покой.")
			os.Exit(0)
		}
	case ERROR_INVALID_CAPTCHA, ERROR_TOO_FAST:
		break
	default:
		post.Log(fmt.Sprintf("неизвестный код = %d; меня пока не научили на это правильно реагировать!", code))
		// TODO: отправлять Message() и Code() как issue.
		// Но для этого нужна сторонняя инфа?
	}
	post.Env.Status <- false
}

// Handle makaba's captcha id response.
func captchaIdErrorHandler(post *Post, cerr *captcha.CaptchaIdError) {
	switch cerr.ErrorId {
	case captcha.CAPTCHA_FAIL:
		post.Log("макаба вернула 0, ошибка получения. Может истекли печенюшки?")
	case captcha.CAPTCHA_HTTP_FAIL: // This can be caused by either 2ch server or proxy.
		if post.Env.UseProxy {
			post.Verbose(cerr.Extra)
		} else {
			post.Log(cerr.Extra)
		}
		post.HTTPFailed++
		post.Log(fmt.Sprintf("%d/%d, не удалось подключиться, ошибка получения капчи.",
			post.HTTPFailed, post.Env.FailedConnectionsLimit))
		if post.HTTPFailed >= post.Env.FailedConnectionsLimit {
			post.Log("прокся исчерпала попытки, удаляю.")
			post.Env.Filter <- post.Proxy.Addr
		}
	default:
		post.Log(fmt.Sprintf("%d неизвестная ошибка при получении капчи!", cerr.ErrorId))
	}
	post.Env.Status <- false
}

// Perform posting steps.
func RunPost(post *Post) {
	failed := func(err error) {
		post.Log(err)
		post.Env.Status <- false
	}
	post.Log("получаю id капчи...")
	cerr := post.SetCaptchaId()
	if cerr != nil {
		post.Log("не смогла получить id капчи.")
		captchaIdErrorHandler(post, cerr)
		return
	}

	post.Log("решаю капчу...")
	var solver captcha.Solver
	switch post.Env.AntiCaptcha {
	case env.RUCAPTCHA:
		solver = captcha.RuCaptchaSolver
	case env.OCR:
		solver = captcha.NeuralSolver
	default:
		solver = captcha.NeuralSolver
	}
	err := post.SolveCaptcha(solver)
	if err != nil {
		failed(err)
		return
	}
	post.Log(fmt.Sprintf("капча решена успешно: %s", post.CaptchaValue))

	params, err := post.MakeParamsMap()
	if err != nil { // This will appear only if we can't get a random thread.
		post.Log("не смогла получить случайный тред.")
		failed(err)
		return
	}
	files := make(map[string][]byte)
	if post.Env.FilesPerPost != 0 {
		files, err = post.MakeFilesMap()
		if err != nil {
			post.Log("не смогла выбрать файлы.")
			failed(err)
			return
		}
	}
	response, err := post.SendPost(params, files)
	if err != nil {
		post.Log("не смогла отправить пост.")
		failed(err)
		return
	}
	post.Log(fmt.Sprintf("%d: %s", response.Code(), response.Message()))
	responseHandler(post, response.Code())
}
