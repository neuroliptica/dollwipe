// engine.go: main posting logic.
// Coherently appling Post's methods and catching errors.

package engine

import (
	"dollwipe/captcha"
	"dollwipe/env"
	"dollwipe/network"
	"fmt"
	"os"
)

func InitPost(penv *env.Env, proxy network.Proxy) *Post {
	post := Post{
		Env:        penv,
		Proxy:      proxy,
		HTTPFailed: 0,
	}
	for key, value := range post.Env.Headers {
		post.Verbose(key, ": ", string(value))
	}
	for i := range post.Env.Cookies {
		post.Verbose("Cookie: ", fmt.Sprintf("%s=%s", post.Env.Cookies[i].Name, post.Env.Cookies[i].Value))
	}
	return &post
}

// Handle makaba's posting response.
func responseHandler(post *Post, code int32) {
	if code == OK {
		post.Env.Status <- true
		return
	}
	switch code {
	case ERROR_BANNED:
		post.Log("прокси забанена, удаляю.")
		post.Env.Filter <- post.Proxy.Addr
	case ERROR_ACCESS_DENIED:
		post.Log("доступ заблокирован, удаляю.")
		post.Env.Filter <- post.Proxy.Addr
	case ERROR_CLOSED:
		post.Log("тред закрыт, маладца.")
		if post.Env.WipeMode == env.SINGLE {
			post.Log("больше не могу постить в этот тред, завершаюсь.")
			os.Exit(0)
		}
	case ERROR_INVALID_CAPTCHA, ERROR_TOO_FAST:
		break
	default:
		post.Log(fmt.Sprintf("неизвестный код = %d; Я пока не знаю, как на это реагировать!", code))
		// TODO: отправлять Message() и Code() как issue.
	}
	post.Env.Status <- false
}

// Handle makaba's captcha id response.
func captchaIdErrorHandler(post *Post, cerr *captcha.CaptchaIdError) {
	switch cerr.ErrorId {
	case captcha.CAPTCHA_FAIL:
		post.Log("макаба вернула 0, ошибка получения. Истекли печенюшки?")
	case captcha.CAPTCHA_HTTP_FAIL: // This can be caused by either 2ch server or proxy.
		if post.Env.UseProxy {
			post.Verbose(cerr.Extra)
		} else {
			post.Log(cerr.Extra)
		}
		post.HTTPFailed++
		post.Log(fmt.Sprintf("%d/%d, ошибка подключения, ошибка получения капчи.",
			post.HTTPFailed, post.Env.FailedConnectionsLimit))
		if post.HTTPFailed >= post.Env.FailedConnectionsLimit {
			post.Log("прокся исчерпала попытки, удаляю.")
			post.Env.Filter <- post.Proxy.Addr
		}
	case captcha.CAPTCHA_NEED_CHECK:
		post.Log("макаба вернула NEED_CHECK. Я пока не знаю, как на это реагировать!")
		post.Log("если вы видите это, то сообщите моему разработчику, пожалуйста!")
	default:
		post.Log(fmt.Sprintf("%d неизвестная ошибка при получении id капчи!", cerr.ErrorId))
	}
	post.Env.Status <- false
}

// Perform posting steps.
// Should be run only when cookies already has been set.
func RunPost(post *Post) {
	failed := func(err error) {
		post.Log(err)
		post.Env.Status <- false
	}
	post.Log("получаю id капчи...")
	cerr := post.SetCaptchaId()
	if cerr != nil {
		post.Log("не удалось получить id капчи.")
		captchaIdErrorHandler(post, cerr)
		return
	}

	post.Log("решаю капчу...")
	err := post.SolveCaptcha(captcha.RuCaptchaSolver)
	if err != nil {
		failed(err)
		return
	}
	post.Log(fmt.Sprintf("капча решена успешно: %s", post.CaptchaValue))

	params, err := post.MakeParamsMap()
	if err != nil { // This will appear only if we can't get a random thread.
		post.Log("не удалось получить случайный тред.")
		failed(err)
		return
	}
	files := make(map[string][]byte)
	if post.Env.FilesPerPost != 0 {
		files, err = post.MakeFilesMap()
		if err != nil {
			post.Log("не удалось выбрать файлы.")
			failed(err)
			return
		}
	}
	response, err := post.SendPost(params, files)
	if err != nil {
		post.Log("не удалось отправить пост.")
		failed(err)
		return
	}
	post.Log(fmt.Sprintf("%d: %s", response.Code(), response.Message()))
	responseHandler(post, response.Code())
}
