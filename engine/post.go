// post.go: main Post struct and it's methods. It's like a single posting unit.
// Post struct encapsulates in itself all posting methods and data for the single post.

package engine

import (
	"dollwipe/captcha"
	"dollwipe/env"
	"dollwipe/network"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/neuroliptica/logger"
)

const (
	OK = iota
)

// Makaba's posting error codes.
const (
	ERROR_TOO_FAST        = -8
	ERROR_CLOSED          = -7
	ERROR_BANNED          = -6
	ERROR_INVALID_CAPTCHA = -5
	ERROR_ACCESS_DENIED   = -4
)

const (
	CAPTCHA_API = "/api/captcha/2chcaptcha/"
	POST_API    = "/user/posting"
)

// General posting API's response interface.
type MakabaResponse interface {
	Code() int32
	Message() string
}

// Posted successfully.
type MakabaOk struct {
	Num    int32
	Result int32
}

// OK = 0, if request is successful.
func (r *MakabaOk) Code() int32 {
	return OK
}

// Get response message.
func (r *MakabaOk) Message() string {
	return fmt.Sprintf("OK, пост %d отправлен.", r.Num)
}

// Posting falied.
type MakabaFail struct {
	Error struct {
		Code    int32
		Message string
	}
	Result int32
}

// Get error code.
func (r *MakabaFail) Code() int32 {
	return r.Error.Code
}

// Get error message.
func (r *MakabaFail) Message() string {
	return r.Error.Message
}

// Single posting unit.
type Post struct {
	Proxy   network.Proxy
	Cookies []*http.Cookie
	Headers map[string]env.Header

	Logger *logger.Logger

	CaptchaId, CaptchaValue string
	Env                     *env.Env
	HTTPFailed              uint64 // Failed requests counter.
}

// General logging purpose method.
func (post *Post) Log(msg ...interface{}) {
	post.Logger.Log(msg...)
}

// Logging with format.
func (post *Post) Logf(format string, msg ...interface{}) {
	post.Logger.Logf(format, msg...)
}

// Extra logs when -v flag is set.
func (post *Post) Verbose(msg ...interface{}) {
	if post.Env.Verbose {
		post.Log(msg...)
	}
}

// Set up custom headers and cookies for posting unit.
func (post *Post) InitPostCookiesAndHeaders() error {
	var err error
	post.Cookies, post.Headers, err = env.GetCookiesAndHeaders(
		post.Proxy,
		time.Second*time.Duration(post.Env.WaitTime),
		post.Verbose)
	return err
}

// Get solver function depending on start params.
func (post *Post) GetCaptchaSolver() captcha.Solver {
	switch post.Env.AntiCaptcha {
	case env.RUCAPTCHA:
		return captcha.RuCaptchaSolver
	case env.OCR:
		return captcha.NeuralSolver
	default:
		return captcha.NeuralSolver
	}
}

// Perform request with post's headers, proxy and cookies.
func (post *Post) PerformReq(req *http.Request) ([]byte, error) {
	// Setting up cookies.
	for i := range post.Cookies {
		req.AddCookie(post.Cookies[i])
	}
	// Setting up headers.
	for key, value := range post.Headers {
		req.Header.Add(key, string(value))
	}
	// Setting up HTTP(s) proxy auth.
	if post.Env.UseProxy && post.Proxy.NeedAuth() && post.Proxy.ProxyType() != "socks" {
		auth := network.MakeProxyAuthHeader(post.Proxy)
		req.Header.Add("Proxy-Authorization", auth)
	}
	transport := network.MakeTransport(post.Proxy)
	if post.Env.UseProxy {
		transport.ProxyConnectHeader = req.Header
	}
	dump, _ := httputil.DumpRequest(req, false)
	post.Verbose(string(dump))

	resp, err := network.PerformReq(req, transport)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	post.Log(resp.Status)

	return network.ReadBody(resp.Body)
}

// Http Get request with post's headers, proxy and cookies.
func (post *Post) SendGet(link string) ([]byte, error) {
	req, err := http.NewRequest("GET", link, nil)
	if err != nil {
		return nil, fmt.Errorf("не удалось сформировать GET запрос к %s", link)
	}
	return post.PerformReq(req)
}

// Get captcha id from the server and set up post.CaptchaId field.
func (post *Post) GetCaptchaId() *captcha.CaptchaIdError {
	link := "https://2ch.hk" + CAPTCHA_API + "id?board=" + post.Env.Board + "&thread=" + strconv.FormatUint(post.Env.Thread, 10)
	cont, err := post.SendGet(link)
	if err != nil {
		post.Log(err)
		cerr := captcha.NewCaptchaIdError(captcha.CAPTCHA_HTTP_FAIL, err)
		return cerr
	}
	var response captcha.CaptchaJSON
	json.Unmarshal(cont, &response)
	if len(response.Id) == 0 {
		return captcha.NewCaptchaIdError(response.Result, nil)
	}
	post.CaptchaId = response.Id
	post.Verbose("Captcha Id Response => ", response)
	return nil
}

// Get captcha image from server by it's id.
func (post *Post) GetCaptchaImage() ([]byte, error) {
	link := "https://2ch.hk" + CAPTCHA_API + "show?id=" + post.CaptchaId
	img, err := post.SendGet(link)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения капчи: %v", err)
	}
	return img, nil
}

// Solver is a function, that must satisfy the signature described in captcha/captcha.go.
func (post *Post) SolveCaptcha(solver captcha.Solver) error {
	img, err := post.GetCaptchaImage()
	if err != nil {
		return fmt.Errorf("ошибка получения капчи: %v", err)
	}
	value, err := solver(img, post.Env.Key)
	if err != nil {
		return fmt.Errorf("ошибка решения капчи: %v", err)
	}
	post.CaptchaValue = captcha.Match(value)
	return nil
}

// Build params map to pass them in multipart request.
func (post *Post) MakeParamsMap() (map[string]string, error) {
	board, thread := post.Env.Board, post.Env.Thread
	if post.Env.WipeMode == env.SHRAPNEL {
		var err error
		thread, err = env.GetRandomThread(board)
		if err != nil {
			return nil, err
		}
	}
	rand.Seed(time.Now().UnixNano())
	params := map[string]string{
		"task":             "post",
		"usercode":         "",
		"code":             "",
		"captcha_type":     "2chcaptcha",
		"oekaki_image":     "",
		"oekaki_metadata":  "",
		"makaka_id":        "",
		"makaka_answer":    "",
		"email":            "",
		"comment":          post.Env.Captions[rand.Intn(len(post.Env.Captions))],
		"board":            board,
		"thread":           strconv.FormatUint(thread, 10),
		"2chcaptcha_id":    post.CaptchaId,
		"2chcaptcha_value": post.CaptchaValue,
	}
	if post.Env.Sage {
		params["email"] = "sage"
	}
	if post.Env.WipeMode == env.CREATING {
		params["thread"] = "0"
	}
	return params, nil
}

// Build files map to pass them in multipart request.
// Total size always will be <= 2 * 10^7 bytes (slightly less, than 20MB).
func (post *Post) MakeFilesMap() (map[string][]byte, error) {
	rand.Seed(time.Now().UnixNano())
	var (
		limit = int(2e7) // size limit in bytes for all files in one post.
		files = make(map[string][]byte)

		l = rand.Intn(len(post.Env.Files))
		n = uint8(0)
	)
	for i := 0; limit > 0 && n != post.Env.FilesPerPost; i++ {
		if i != 0 && (l+i)%len(post.Env.Files) == l {
			break
		}
		file := post.Env.Files[(l+i)%len(post.Env.Files)]
		cont := file.Content
		if post.Env.Colorize {
			cont = file.Colorize()
		}
		if len(cont) > limit {
			continue
		}
		limit -= len(cont)
		files[file.RandName()] = cont
		n++
	}
	if n == 0 {
		return nil, fmt.Errorf("все файлы превышают допустимый размер.")
	}
	if n != post.Env.FilesPerPost {
		post.Logf("%d/%d файлов будет прикреплено, суммарный размер превышает 20МБ.",
			n, post.Env.FilesPerPost)
	}
	return files, nil
}

// Sending post with body containing params and files.
func (post *Post) SendPost(params map[string]string, files map[string][]byte) (MakabaResponse, error) {
	var (
		link = "https://2ch." + post.Env.Domain + POST_API
		ok   MakabaOk
		fail MakabaFail
		form = network.FileForm{"file[]", files}
	)
	req, err := network.NewPostMultipartRequest(link, params, &form)
	if err != nil {
		return nil, fmt.Errorf("не удалось сформировать запрос: %v", err)
	}
	cont, err := post.PerformReq(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	post.Verbose("Makaba Response => ", string(cont))

	json.Unmarshal(cont, &ok)
	if ok.Num == 0 {
		json.Unmarshal(cont, &fail)
		if fail.Error.Code == 0 && fail.Error.Message == "" {
			return nil, fmt.Errorf("сервер вернул неожиданный ответ: %s.", string(cont))
		} else if fail.Error.Code == 0 && fail.Error.Message != "" {
			// for some reason there is non empty ban with code 0
			fail.Error.Code = ERROR_BANNED
		}
		return &fail, nil
	}
	return &ok, nil
}
