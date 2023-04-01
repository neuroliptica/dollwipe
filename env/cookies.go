// cookies.go: bypass cloudflare, get cookies and headers for future requests.

package env

import (
	"net/http"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

// Cast proto.NetworkCookie to http.Cookie.
func protoToHttp(pCookies []*proto.NetworkCookie) []*http.Cookie {
	cookies := make([]*http.Cookie, 0)
	for i := range pCookies {
		cookie := http.Cookie{Name: pCookies[i].Name, Value: pCookies[i].Value}
		cookies = append(cookies, &cookie)
	}
	return cookies
}

// TODO: extract request headers.
// TODO: make request optionally using proxy.
// Create browser instance, pass cloudflare, get cookies and headers.
func GetHeaders(url string, wait time.Duration) ([]*http.Cookie, map[string]Header) {
	browser := rod.New().Timeout(time.Minute).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("")

	var e proto.NetworkRequestWillBeSent

	// Passing Cloudflare
	page.MustNavigate(url)
	page.MustWaitNavigation()
	time.Sleep(wait)

	// Request to the captcha -> 200; then extract headers and cookies
	waitRequest := page.WaitEvent(&e)
	page.MustNavigate("https://2ch.hk/api/captcha/2chcaptcha/id?board=b&thread=0")
	page.MustWaitNavigation()
	// time.Sleep(wait)
	//var s string
	//fmt.Scan(&s)
	waitRequest()

	//fmt.Println(utils.Dump(
	//	e.Request.URL,
	//	e.Request.Headers))

	cookies := page.MustCookies(url)
	headers := make(map[string]Header, 0)
	//for key, value := range e.Request.Headers {
	//	headers[key] = Header(value.String())
	//}
	headers["Accept"] = Header("text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9")
	headers["Accept-Encoding"] = Header("gzip, deflate, br")
	headers["Accept-Language"] = Header("en")
	headers["Sec-Fetch-Dest"] = Header("document")
	headers["Sec-Fetch-Mode"] = Header("navigate")
	headers["Sec-Fetch-Site"] = Header("none")
	headers["Sec-Fetch-User"] = Header("?1")
	headers["Upgrade-Insecure-Requests"] = Header("1")
	headers["User-Agent"] = Header("Mozilla/5.0 (Macintosh; Intel Mac OS X 11_0_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")
	return protoToHttp(cookies), headers
}
