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
	samesite := map[string]http.SameSite{
		"None":   3,
		"Strict": 2,
		"Lax":    1,
	}
	for i := range pCookies {
		cookie := http.Cookie{
			Name:     pCookies[i].Name,
			Value:    pCookies[i].Value,
			HttpOnly: pCookies[i].HTTPOnly,
			Secure:   pCookies[i].Secure,
			Domain:   pCookies[i].Domain,
			Path:     pCookies[i].Path,
			Expires:  pCookies[i].Expires.Time(),
		}
		if val, ok := samesite[string(pCookies[i].SameSite)]; ok {
			cookie.SameSite = val
		}
		cookies = append(cookies, &cookie)
	}
	return cookies
}

// TODO: proxy support.
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

	// Request to the captcha -> 200; then extract cookies.
	waitRequest := page.WaitEvent(&e)
	page.MustNavigate("https://2ch.hk/api/captcha/2chcaptcha/id?board=b&thread=0")
	page.MustWaitNavigation()
	waitRequest()

	cookies := page.MustCookies(url)
	headers := make(map[string]Header, 0)

	// Manually set up, because they won't change.
	headers["Accept"] = Header("text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	headers["Accept-Encoding"] = Header("gzip, deflate, br")
	headers["Accept-Language"] = Header("en-US,en;q=0.5")
	headers["DNT"] = Header("1")
	headers["Sec-Fetch-Dest"] = Header("document")
	headers["Sec-Fetch-Mode"] = Header("navigate")
	headers["Sec-Fetch-Site"] = Header("none")
	headers["Sec-Fetch-User"] = Header("?1")
	headers["Upgrade-Insecure-Requests"] = Header("1")
	headers["User-Agent"] = Header("Mozilla/5.0 (Macintosh; Intel Mac OS X 11_0_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36")

	return protoToHttp(cookies), headers
}
