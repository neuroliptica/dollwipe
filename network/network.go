// network.go: helping network utils, which separated from Post.
// E.g. create new request, new multipart request, etc.

package network

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"time"
)

// Struct for passing files in multipart form.
// FormName is corresponding to <name=FormName>
// Files is corresponding to pairs of (filename, content).
type FileForm struct {
	FormName string
	Files    map[string][]byte
}

// Single proxy unit, on which single Post instance will be based.
type Proxy struct {
	Addr       string // For logging purpose.
	AddrParsed *url.URL

	Login, Pass string // If proxy is public, then these fields will be empty.
}

func (p Proxy) NoProxy() bool {
	return p.Addr == "localhost"
}

func (p Proxy) String() string {
	return p.Addr
}

// Build new HTTP POST request to link with params in query.
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

// Build new HTTP POST request with multipart-form data body.
// Files should be passed as the single FileForm instance pointer.
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

// Build HTTP GET request, perform it and return response body.
func SendGet(link string) ([]byte, error) {
	resp, err := http.Get(link)
	if err != nil {
		return make([]byte, 0), err
	}
	defer resp.Body.Close()
	cont, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return make([]byte, 0), err
	}
	return cont, nil
}

// Extra transport pointer should be passed if we want to set up proxy.
// Otherwise use nil.
func PerformReq(req *http.Request, transport *http.Transport) (*http.Response, error) {
	client := &http.Client{
		Timeout: time.Second * 20,
	}
	if transport != nil {
		client.Transport = transport
	}
	return client.Do(req)
}
