// post.go: main Post struct and it's methods.
// Post struct encapsulates in itself all posting methods and data for the single post.

package engine

import (
	"crypto/tls"
	"dollwipe/captcha"
	"dollwipe/env"
	"dollwipe/network"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"
)

const (
	OK = iota
	BANNED
	FAILED
	SOLVER_FAILED
)

const (
	ERROR_CLOSED          = -7
	ERROR_BANNED          = -6
	ERROR_INVALID_CAPTCHA = -5
	ERROR_ACCESS_DENIED   = -4

	CAPTCHA_API = "/api/captcha/2chcaptcha/"
	POST_API    = "/makaba/posting.fcgi?json=1"
)

var userAgents = []string{
	"Mozilla/5.0 (X11; Linux x86_64; rv:91.0) Gecko/20100101 Firefox/91.0",
}

// General posting API's response interface.
type MakabaResponse interface {
	Code() int32
	Message() string
}

// Posted successfully.
type MakabaOk struct {
	Status string
	Num    int32
}

func (r *MakabaOk) Code() int32 {
	return OK
}

func (r *MakabaOk) Message() string {
	var msg string
	if r.Status == "Redirect" {
		msg = "OK, тред создан."
	} else {
		msg = fmt.Sprintf("OK, пост %d отправлен.", r.Num)
	}
	return msg
}

// Posting falied.
type MakabaFail struct {
	ErrorCode int32 `json:"Error"`
	Reason    string
}

func (r *MakabaFail) Code() int32 {
	return r.ErrorCode
}

func (r *MakabaFail) Message() string {
	return r.Reason
}

// Single posting unit.
type Post struct {
	Proxy     network.Proxy
	Cookie    []*http.Cookie
	UserAgent string

	CaptchaId, CaptchaValue string
	Env                     *env.Env
}

func (post *Post) Log(msg ...interface{}) {
	post.Env.Logger <- fmt.Sprintf("%s %s",
		post.Proxy.String(), fmt.Sprint(msg...))
}

func (post *Post) Verbose(msg ...interface{}) {
	if !post.Env.Verbose {
		return
	}
	post.Log(msg...)
}

// Custom TLS transport for send requests with proxy.
func (post *Post) MakeTransport() *http.Transport {
	config := &tls.Config{
		InsecureSkipVerify: true,
	}
	proto := make(map[string]func(string, *tls.Conn) http.RoundTripper)

	transport := &http.Transport{
		Proxy:           http.ProxyURL(post.Proxy.AddrParsed),
		TLSClientConfig: config,
		TLSNextProto:    proto,
	}
	return transport
}

// Perform request with post headers.
func (post *Post) PerformReq(req *http.Request) ([]byte, error) {
	for i := range post.Env.Cookies {
		req.AddCookie(post.Env.Cookies[i])
	}
	req.Header.Add("User-Agent", post.UserAgent)
	if post.Proxy.Login != "" && post.Proxy.Pass != "" {
		credits := post.Proxy.Login + ":" + post.Proxy.Pass
		basicAuth := "Basic " + base64.StdEncoding.EncodeToString([]byte(credits))
		req.Header.Add("Proxy-Authorization", basicAuth)
	}
	//post.Verbose(req)

	// Setting up proxy.
	var transport *http.Transport
	if post.Env.UseProxy {
		transport = post.MakeTransport()
	}
	resp, err := network.PerformReq(req, transport)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	post.Log(resp.Status)
	cont, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return cont, nil
}

func (post *Post) SendGet(link string) ([]byte, error) {
	req, err := http.NewRequest("GET", link, nil)
	post.Verbose("отправляю HTTP GET запрос на ", link)
	if err != nil {
		return nil, fmt.Errorf("не удалось сформировать GET запрос к %s", link)
	}
	return post.PerformReq(req)
}

func (post *Post) SetUserAgent() {
	rand.Seed(time.Now().UnixNano())
	post.UserAgent = userAgents[rand.Intn(len(userAgents))]
	post.Verbose("USER-AGENT = ", post.UserAgent)
}

func (post *Post) SetPasscode() {
	return
}

// Get captcha id from 2ch server and set CaptchaId field.
func (post *Post) SetCaptchaId() *captcha.CaptchaIdError {
	link := "https://2ch." + post.Env.Domain + CAPTCHA_API + "id"
	cont, err := post.SendGet(link)
	if err != nil {
		cerr := captcha.NewCaptchaIdError(captcha.CAPTCHA_HTTP_FAIL, err)
		return cerr
	}
	var response captcha.CaptchaJSON
	json.Unmarshal(cont, &response)
	if len(response.Id) == 0 {
		return captcha.NewCaptchaIdError(response.Result, nil)
	}
	post.CaptchaId = response.Id
	post.Verbose("CAPTCHA ID RESPONSE = ", response)
	return nil
}

// Get captcha image from server by it's id.
func (post *Post) GetCaptchaImage() ([]byte, error) {
	link := "https://2ch." + post.Env.Domain + CAPTCHA_API + "show?id=" + post.CaptchaId
	img, err := post.SendGet(link)
	if err != nil {
		return nil, fmt.Errorf("ошибка получения капчи: %v", err)
	}
	return img, nil
}

func (post *Post) SolveCaptcha(solver captcha.Solver) error {
	img, err := post.GetCaptchaImage()
	if err != nil {
		return err
	}
	value, err := solver(img, post.Env.Key)
	if err != nil {
		return fmt.Errorf("ошибка решения капчи: %v", err)
	}
	post.CaptchaValue = value
	return nil
}

// Map data for basic form in multipart request.
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
		"thread":           thread,
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

// Map data for file form in multipart request.
func (post *Post) MakeFilesMap() (map[string][]byte, error) {
	rand.Seed(time.Now().UnixNano())
	var (
		limit = int(2e7) // size limit in bytes for all files in one post.
		files = make(map[string][]byte)

		l = rand.Intn(len(post.Env.Files))
		n = uint8(0)
	)
	// Using greedy, in worst case O(n), where n -- post.Env.Files size.
	for i := 0; limit > 0 && n != post.Env.FilesPerPost; i++ {
		if i != 0 && (l+i)%len(post.Env.Files) == l {
			break
		}
		file := post.Env.Files[(l+i)%len(post.Env.Files)]
		if len(file.Content) > limit {
			continue
		}
		limit -= len(file.Content)
		files[file.RandName()] = file.Content
		n++
	}
	if n == 0 {
		return nil, fmt.Errorf("все файлы превышают допустимый размер.")
	}
	if n != post.Env.FilesPerPost {
		post.Log(fmt.Sprintf("%d/%d файлов будет прикреплено, суммарный размер превышает 20МБ.",
			n, post.Env.FilesPerPost))
	}
	return files, nil
}

func (post *Post) SendPost(params map[string]string, files map[string][]byte) (MakabaResponse, error) {
	var (
		link = "https://2ch." + post.Env.Domain + POST_API
		ok   MakabaOk
		fail MakabaFail
		form = network.FileForm{"formimages[]", files}
	)
	req, err := network.NewPostMultipartRequest(link, params, &form)
	if err != nil {
		return nil, fmt.Errorf("не удалось сформировать запрос: %v", err)
	}
	cont, err := post.PerformReq(req)
	if err != nil {
		return nil, fmt.Errorf("ошибка отправки запроса: %v", err)
	}
	post.Verbose("MAKABA RESPONSE = ", string(cont))
	json.Unmarshal(cont, &ok)
	if ok.Status == "" {
		json.Unmarshal(cont, &fail)
		if fail.ErrorCode == 0 {
			return nil, fmt.Errorf("сервер вернул неожиданный ответ.")
		}
		return &fail, nil
	}
	return &ok, nil
}
