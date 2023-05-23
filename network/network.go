// network.go: helping network utils, which separated from Post.
// E.g. create new request, new multipart request, net TLS transport, etc.

package network

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/neuroliptica/logger"
	"golang.org/x/net/proxy"
)

var CheckerLogger = logger.MakeLogger("checker").BindToDefault()

// Struct for passing files in multipart form.
// FormName is corresponding to <name=FormName>
// Files is corresponding to the pairs of (filename, content).
type FileForm struct {
	FormName string
	Files    map[string][]byte
}

// Single proxy unit, on which single Post instance will be based.
type Proxy struct {
	Addr       string // For logging purpose.
	AddrParsed *url.URL

	Login, Pass string // If proxy is public, then these fields will be empty.
	Protocol    string

	SessionId int // Usefull only with -s flag.
	Alive     bool
}

// Default value indicates no proxy.
func (p Proxy) NoProxy() bool {
	return p.Addr == "localhost"
}

// Check if auth is need.
func (p Proxy) NeedAuth() bool {
	return p.Login != "" && p.Pass != ""
}

// Separate Http(s) and Socks proxies.
func (p Proxy) ProxyType() string {
	switch p.Protocol {
	case "http", "https":
		return p.Protocol
	default:
		return "socks"
	}
}

// Stringer interface instance.
func (p Proxy) String() string {
	if p.Addr == "localhost" {
		return p.Addr
	}
	return strings.Split(p.Addr, "//")[1]
}

// Logging purpose, if -s flag is set.
func (p Proxy) StringSid() string {
	if p.SessionId > 0 {
		return fmt.Sprintf("%s[sid=%d]", p, p.SessionId)
	}
	return p.String()
}

// Check if proxy (both socks and http(s)) is alive.
func (p *Proxy) CheckAlive(timeout time.Duration) {
	if p.NoProxy() {
		CheckerLogger.Logf("[%s] => ok", p.String())
		p.Alive = true
		return
	}
	transport := MakeTransport(*p)
	req, err := http.NewRequest("GET", "https://api.ipify.org?format=json", nil)
	if err != nil {
		CheckerLogger.Logf("[%s] => error, internal error: %v", p.String(), err)
		p.Alive = false
		return
	}
	if p.NeedAuth() && p.ProxyType() != "socks" {
		auth := MakeProxyAuthHeader(*p)
		req.Header.Add("Proxy-Authorization", auth)
	}
	transport.ProxyConnectHeader = req.Header
	client := &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		CheckerLogger.Logf("[%s] => error: %v", p.String(), err)
		p.Alive = false
		return
	}
	resp.Body.Close()
	CheckerLogger.Logf("[%s] => ok", p.String())
	p.Alive = true
}

// Create base64 Proxy-Authorization header value.
func MakeProxyAuthHeader(p Proxy) string {
	credits := p.Login + ":" + p.Pass
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(credits))
}

// Build custom TLS transport for sending requests with proxy.
func MakeTransport(p Proxy) *http.Transport {
	config := &tls.Config{
		MaxVersion:         tls.VersionTLS13,
		InsecureSkipVerify: true,
	}
	if p.NoProxy() {
		return &http.Transport{
			TLSClientConfig: config,
			//DisableCompression: true,
		}
	}
	proto := make(map[string]func(string, *tls.Conn) http.RoundTripper)
	transport := &http.Transport{
		TLSClientConfig:   config,
		TLSNextProto:      proto,
		DisableKeepAlives: true,
	}
	// Setting up socks proxy.
	if p.ProxyType() == "socks" {
		auth := &proxy.Auth{
			User:     p.Login,
			Password: p.Pass,
		}
		if p.Protocol == "socks4" || !p.NeedAuth() {
			auth = nil
		}
		dialer, _ := proxy.SOCKS5("tcp", p.String(), auth, proxy.Direct)
		//dialer, _ := proxy.SOCKS5("tcp", p.String(), auth, &net.Dialer{
		//	Timeout:   time.Second * 5,
		//	KeepAlive: time.Second * 5,
		//})
		transport.Dial = dialer.Dial
	} else {
		transport.Proxy = http.ProxyURL(p.AddrParsed)
	}
	return transport
}

// Build new Http Post request to the link with params in query.
func NewPostRequest(link string, params map[string]string) (*http.Request, error) {
	query := url.Values{}
	for key, value := range params {
		query.Add(key, value)
	}
	req, err := http.NewRequest("POST", link+query.Encode(), nil)
	if err != nil {
		return nil, err
	}
	return req, nil
}

// Build new Http Post request with multipart-form data body.
func NewPostMultipartRequest(link string, params map[string]string, files *FileForm) (*http.Request, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for key, value := range params {
		err := writer.WriteField(key, value)
		if err != nil {
			writer.Close()
			return nil, err
		}
	}
	for file, cont := range files.Files {
		part, err := writer.CreateFormFile(files.FormName, file)
		if err != nil {
			writer.Close()
			return nil, err
		}
		part.Write(cont)
	}
	writer.Close()

	req, err := http.NewRequest("POST", link, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return req, nil
}

// Read response body.
func ReadBody(body io.Reader) ([]byte, error) {
	cont, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}
	return cont, nil
}

// Build Http Get request, perform it and return response.
func SendGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ReadBody(resp.Body)
}

// Build Http Post request with payload in body, perform it and return response.
func SendPost(url string, payload []byte, headers map[string]string, transport *http.Transport) ([]byte, error) {
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := PerformReq(req, transport)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return ReadBody(resp.Body)
}

// Perform request using transport, transport can be nil if not required.
func PerformReq(req *http.Request, transport *http.Transport) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 30,
	}
	if transport != nil {
		client.Transport = transport
	}
	return client.Do(req)
}
