package captcha

import (
	"dollwipe/network"
	"encoding/base64"
	"encoding/json"
)

const ManualUrl = "http://localhost:8080"

type (
	// Error body response schema.
	ResponseErrorBody struct {
		Failed   bool   `json:"failed"`
		Reason   string `json:"reason"`
		Internal string `json:"internal"`
	}
	//
	// ["error"]["reason"] description:
	//
	//	INVALID_BODY =>
	//		failed to provide data to server;
	//		can be caused either by invalid Content-Type header
	//		or invalid fileds in request json payload.
	//
	//	QUEUED =>
	//		captcha was requested by /get method and is currently
	//		solving by client.
	//		should wait a bit.
	//
	//	UNSOLVED =>
	//		captcha yet wasn't requested by /get method;
	//		should wait a bit (90 sec at least), then abort.
	//
	//	FAILED =>
	//		captcha failed to solve (either timeout or invalid hash);
	//		should abort.
	//

	// POST /post request schema.
	PostRequestBody struct {
		Data string `json:"data"`
	}

	// POST /post response schema.
	PostResponseBody struct {
		Hash  string            `json:"hash"`
		Error ResponseErrorBody `json:"error"`
	}

	CheckRequestBody struct {
		Hash string `json:"hash"`
	}

	CheckResponseBody struct {
		Value string            `json:"value"`
		Error ResponseErrorBody `json:"error"`
	}
)

// Make ResponseErrorBody instance of error interface.
func (e ResponseErrorBody) Error() string {
	return e.Reason
}

// Call /post API method, post captcha on server and return hash.
func PostCaptcha(img []byte) (string, error) {
	body := PostRequestBody{
		Data: "data:image/png;base64," + base64.StdEncoding.EncodeToString(img),
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	cont, err := network.SendPost(
		ManualUrl+"/post",
		payload,
		map[string]string{
			"Content-Type": "application/json",
		},
		nil)
	if err != nil {
		return "", err
	}
	var resp PostResponseBody
	json.Unmarshal(cont, &resp)
	if resp.Hash == "" || resp.Error.Failed {
		return "", resp.Error
	}
	return resp.Hash, nil
}

// Call /check API method, check captcha status.
func CheckCaptcha(hash string) (string, error) {
	body := CheckRequestBody{
		Hash: hash,
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	cont, err := network.SendPost(
		ManualUrl+"/check",
		payload,
		map[string]string{
			"Content-Type": "application/json",
		},
		nil)
	if err != nil {
		return "", nil
	}
	var resp CheckResponseBody
	json.Unmarshal(cont, &resp)
	if resp.Value == "" || resp.Error.Failed {
		return "", resp.Error
	}
	return resp.Value, nil
}
