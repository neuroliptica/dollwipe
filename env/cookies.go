// cookies.go: bypass cloudflare, get cookies and headers for future requests.

package env

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
	"github.com/go-rod/rod/lib/utils"
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
// Create browser instance, pass cloudflare, get cookies and headers.
func ParseCookies(url string, wait time.Duration) ([]*http.Cookie, map[string]Header) {
	browser := rod.New().Timeout(time.Minute).MustConnect()
	defer browser.MustClose()

	page := browser.MustPage("")

	var e proto.NetworkRequestWillBeSent

	// Passing Cloudflare
	page.MustNavigate(url)
	page.MustWaitNavigation()
	time.Sleep(wait)

	// Request to the captcha -> 200; then extract headers and cookies
	wait_ := page.WaitEvent(&e)
	page.MustNavigate("https://2ch.hk/api/captcha/2chcaptcha/id?board=b&thread=0")
	page.MustWaitNavigation()
	time.Sleep(wait)
	wait_()

	fmt.Println(utils.Dump(
		e.Request.URL,
		e.Request.Headers))

	cookies := page.MustCookies(url)
	headers := make(map[string]Header, 0)
	for key, value := range e.Request.Headers {
		headers[key] = Header(value.String())
	}
	return protoToHttp(cookies), headers
}
