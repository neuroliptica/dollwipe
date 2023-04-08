// network.go: helping network utils, which separated from Post.
// E.g. create new request, new multipart request, net TLS transport, etc.

package network

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/proxy"
)

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
		TLSClientConfig: config,
		TLSNextProto:    proto,
	}
	// Setting up socks proxy.
	if p.ProxyType() == "socks" {
		auth := &proxy.Auth{
			User:     p.Login,
			Password: p.Pass,
		}
		if p.Protocol == "socks4" {
			auth = nil
		}
		dialer, _ := proxy.SOCKS5("tcp", p.String(), auth, proxy.Direct)
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

// Build Http Get request, perform it and return response body.
func SendGet(link string) ([]byte, error) {
	resp, err := http.Get(link)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	cont, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return cont, nil
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
