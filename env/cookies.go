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
	"math/rand"
	"net/http"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

const (
	captchaApi = "https://2ch.hk/api/captcha/2chcaptcha/id?board=b&thread=0"
	mainPage   = "https://2ch.hk/b"
)

// Value type for header pair.
type Header string

// Logging callback function signature.
type LogCallback = func(...interface{})

// Screen Structure
type ScreenStructure struct {
	Width  int
	Height int
}

// Gen Random Device Screen
func RandomDeviceScreen() ScreenStructure {
	Screens := [7]ScreenStructure{
		{1366, 768},
		{1920, 1080},
		{1280, 1024},
		{1600, 900},
		{1380, 800},
		{1024, 768},
		{1440, 900},
	}

	return Screens[rand.Intn(len(Screens))]
}

// Gen Random Device Pixel Ratio
func RandomDevicePixelRatio() float64 {
	PixelRatios := [3]float64{
		1,
		1.25,
		1.5,
	}

	return PixelRatios[rand.Intn(len(PixelRatios))]
}

// Gen Random Device
func RandomDevice(UserAgent string) devices.Device {
	Screen := RandomDeviceScreen()

	return devices.Device{
		Title:          "Windows",
		Capabilities:   []string{},
		UserAgent:      UserAgent,
		AcceptLanguage: "en",
		Screen: devices.Screen{
			DevicePixelRatio: RandomDevicePixelRatio(),
			Horizontal: devices.ScreenSize{
				Width:  Screen.Width,
				Height: Screen.Height,
			},
			Vertical: devices.ScreenSize{
				Width:  Screen.Height,
				Height: Screen.Width,
			},
		},
	}
}

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
func MakeRequestWithMiddleware(p network.Proxy, wait time.Duration, logger LogCallback) (cookies []*http.Cookie, err error) {
	defer func() {
		if r := recover(); r != nil {
			logger("[rod-debug] panic!: %v", r)
			err = fmt.Errorf("возникла внутренняя ошибка")
		}
	}()

	u := launcher.New().
		Set("--force-webrtc-ip-handling-policy", "disable_non_proxied_udp").
		Set("--enforce-webrtc-ip-permission-check", "False").
		Set("--use-gl", "osmesa").
		MustLaunch()

	browser := rod.New().ControlURL(u).Timeout(5 * time.Minute).MustConnect()
	defer browser.Close()

	page := browser.MustPage("")
	Device := RandomDevice("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.63 Safari/537.36")
	page.MustEmulate(Device)
	page.MustSetViewport(Device.Screen.Horizontal.Width, Device.Screen.Horizontal.Height, 0, false)
	page.MustSetExtraHeaders("cache-control", "max-age=0")
	page.MustSetExtraHeaders("sec-ch-ua", `Google Chrome";v="102", "Chromium";v="102", ";Not A Brand";v="102"`)
	page.MustSetExtraHeaders("sec-fetch-site", "same-origin")
	page.MustSetExtraHeaders("sec-fetch-user", "?1")
	page.SetUserAgent(&proto.NetworkSetUserAgentOverride{
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.63 Safari/537.36",
		AcceptLanguage: "ru-RU,ru;=0.9",
		Platform:       "Windows",
	})
	page.MustEvalOnNewDocument(`localStorage.clear();`)

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
		logger(fmt.Sprintf("[rod-debug] [%d] %s",
			ctx.Response.Payload().ResponseCode, ctx.Request.URL()))
		ctx.LoadResponse(&client, true)
	})
	go router.Run()

	err = page.Navigate(mainPage)
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

	cookie, err := page.Cookies([]string{captchaApi})
	if err != nil {
		return nil, err
	}
	return protoToHttp(cookie), nil
}

// Create browser instance, pass cloudflare, get cookies and headers.
func GetCookiesAndHeaders(p network.Proxy, wait time.Duration, logger LogCallback) ([]*http.Cookie, map[string]Header, error) {
	cookies, err := MakeRequestWithMiddleware(p, wait, logger)
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
		"User-Agent":                Header("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.63 Safari/537.36"),
	}
	return cookies, headers, nil
}
