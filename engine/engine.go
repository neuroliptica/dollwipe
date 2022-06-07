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

// Init this one in main.
// Proxy full path is a key.
var Posts map[string]*Post

func InitPost(penv *env.Env, proxy network.Proxy) *Post {
	post := Post{
		Env:   penv,
		Proxy: proxy,
	}
	post.SetUserAgent()
	for i := range post.Env.Cookies {
		post.Verbose("COOKIE = ", fmt.Sprintf("%s=%s", post.Env.Cookies[i].Name, post.Env.Cookies[i].Value))
	}

	return &post
}

func responseHandler(post *Post, code int32) {
	if code == OK {
		post.Env.Status <- true
		return
	}
	switch code {
	case ERROR_BANNED:
		post.Log("прокси забанена, удаляю.")
		delete(Posts, post.Proxy.Addr)
	case ERROR_ACCESS_DENIED:
		post.Log("доступ заблокирован, удаляю.")
		delete(Posts, post.Proxy.Addr)
	case ERROR_CLOSED:
		post.Log("тред закрыт, маладца.")
		if post.Env.WipeMode == env.SINGLE {
			post.Log("больше не могу постить в этот тред, завершаюсь.")
			os.Exit(0)
		}
	case ERROR_INVALID_CAPTCHA:
		break
	default:
		post.Log(fmt.Sprintf("неизвестный код = %d; Я пока не знаю, как на это реагировать!", code))
		// TODO: отправлять Message() и Code() как issue.
	}
	post.Env.Status <- false
}

func captchaIdErrorHandler(post *Post, cerr *captcha.CaptchaIdError) {
	switch cerr.ErrorId {
	case captcha.CAPTCHA_FAIL:
		post.Log("макаба вернула 0, ошибка получения. Истекли печенюшки?")
	case captcha.CAPTCHA_HTTP_FAIL:
		post.Log(cerr.Extra)
		post.Log("ошибка отправки запроса. Кажется, плохая прокся. Удаляю.")
		delete(Posts, post.Proxy.Addr)
	case captcha.CAPTCHA_NEED_CHECK:
		post.Log("макаба вернула NEED_CHECK. Я пока не знаю, как на это реагировать!")
		post.Log("если вы видите это, то сообщите моему разработчику, пожалуйста!")
	default:
		post.Log(fmt.Sprintf("%d неизвестная ошибка при получении id капчи!",
			cerr.ErrorId))
	}
	post.Env.Status <- false
}

// Run when cookies is already has been set.
func RunPost(post *Post) {
	failed := func(err error) {
		post.Log(err)
		post.Env.Status <- false
	}

	post.Log("получаю id капчи.")
	cerr := post.SetCaptchaId()
	if cerr != nil {
		post.Log("не удалось получить id капчи!")
		captchaIdErrorHandler(post, cerr)
		return
	}

	post.Log("решаю капчу.")
	err := post.SolveCaptcha(captcha.RuCaptchaSolver)
	if err != nil {
		failed(err)
		return
	}
	post.Log(fmt.Sprintf("капча решена успешно: %s", post.CaptchaValue))

	params, err := post.MakeParamsMap()
	if err != nil { // This will appear only if we can't get a random thread.
		post.Log("не удалось получить случайный тред!")
		failed(err)
		return
	}
	files := make(map[string][]byte)
	if post.Env.FilesPerPost != 0 {
		files, err = post.MakeFilesMap()
		if err != nil {
			post.Log("не удалось выбрать файлы!")
			failed(err)
			return
		}
	}
	response, err := post.SendPost(params, files)
	if err != nil {
		post.Log("не удалось отправить пост!")
		failed(err)
		return
	}
	post.Log(fmt.Sprintf("%d: %s", response.Code(), response.Message()))
	responseHandler(post, response.Code())
}
