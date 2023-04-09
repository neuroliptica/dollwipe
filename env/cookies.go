// cookies.go: bypass cloudflare, get cookies and headers for future requests.
//
// Bypass scheme looks like this.
// REQ - request, RESP - response, MID - middleware.
//
//		REQ 2ch.hk/b -> MID (set up proxy) -> SERVER -> wait until cloudflare finished
//		-> RESP -> MID (response unmodified) -> CLIENT
//
//		After this, we should get "Set-Cookies" header already.
//		To check if cookies has set up process one more chain:
//
//		REQ 2ch.hk/api/captcha/... -> MID (set up proxy) -> SERVER -> wait until navigation
//		-> RESP -> MID (response unmodified) -> CLIENT
//
//		After chain finished, we can extract request cookies and finally use them.

package env

import (
	"dollwipe/network"
	"fmt"
	"net/http"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/proto"
)

const (
	captchaApi = "https://2ch.hk/api/captcha/2chcaptcha/id?board=b&thread=0"
	mainPage   = "https://2ch.hk/b"
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

// Create webdriver instance and pass it's requests through custom middleware.
// Error will be returned if some of the requests has failed. Otherwise re-
// turn value should be processed as a successful, even if it is empty.
func MakeRequestWithMiddleware(p network.Proxy, wait time.Duration) ([]*http.Cookie, error) {
	browser := rod.New().Timeout(2 * time.Minute).MustConnect()
	defer browser.Close()

	page := browser.MustPage("")
	router := page.HijackRequests()
	defer router.Stop()

	// Do not send shit reuqests for faster bypass.
	router.MustAdd("*.jpg", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonAborted)
	})
	router.MustAdd("*.gif", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonAborted)
	})
	router.MustAdd("*.png", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonAborted)
	})
	router.MustAdd("*google*", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonAborted)
	})
	router.MustAdd("*24smi*", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonAborted)
	})
	router.MustAdd("*yadro.ru*", func(ctx *rod.Hijack) {
		ctx.Response.Fail(proto.NetworkErrorReasonAborted)
	})

	// When request is hijacked, custom trasport will be set.
	// For http(s) proxies with authorization will set an auth header.
	// Hijacked response will return unmodified.
	router.MustAdd("*", func(ctx *rod.Hijack) {
		transport := network.MakeTransport(p)
		if p.ProxyType() != "socks" && p.NeedAuth() {
			auth := network.MakeProxyAuthHeader(p)
			ctx.Request.Req().Header.Set("Proxy-Authorization", auth)
		}
		if !p.NoProxy() {
			transport.ProxyConnectHeader = ctx.Request.Req().Header
		}
		client := http.Client{
			Transport: transport,
			Timeout:   2 * time.Minute,
		}
		fmt.Println(ctx.Request.URL())
		ctx.LoadResponse(&client, true)
	})
	go router.Run()

	err := page.Navigate(mainPage)
	if err != nil {
		return nil, err
	}
	page.MustWaitNavigation()
	time.Sleep(wait)

	err = page.Navigate(captchaApi)
	if err != nil {
		return nil, err
	}
	page.MustWaitLoad()

	cookies, err := page.Cookies([]string{captchaApi})
	if err != nil {
		return nil, err
	}
	return protoToHttp(cookies), nil
}

// Create browser instance, pass cloudflare, get cookies and headers.
func GetCookiesAndHeaders(p network.Proxy, wait time.Duration) ([]*http.Cookie, map[string]Header, error) {
	cookies, err := MakeRequestWithMiddleware(p, wait)
	if err != nil {
		return nil, nil, err
	}
	headers := map[string]Header{
		"Accept":                    Header("text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8"),
		"Accept-Language":           Header("en-US,en;q=0.5"),
		"DNT":                       Header("1"),
		"Sec-Fetch-Dest":            Header("document"),
		"Sec-Fetch-Mode":            Header("navigate"),
		"Sec-Fetch-Site":            Header("none"),
		"Sec-Fetch-User":            Header("?1"),
		"Upgrade-Insecure-Requests": Header("1"),
		"User-Agent":                Header("Mozilla/5.0 (Macintosh; Intel Mac OS X 11_0_0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/87.0.4280.88 Safari/537.36"),
	}
	return cookies, headers, nil
}
