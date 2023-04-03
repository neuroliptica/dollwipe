package captcha

import (
	"bytes"
	"dollwipe/network"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type RequestBody struct {
	Data []string `json:"data"`
}

type ResponseOk struct {
	Data          []string
	Is_generating bool
	Duration      float64
}

const url = "http://127.0.0.1:7860/api/predict"

func NeuralSolver(img []byte, key string) (string, error) {
	body := RequestBody{
		Data: []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(img)},
	}
	bodyBytes, err := json.Marshal(body)
	//log.Println(string(bodyBytes))
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")
	if err != nil {
		return "", err
	}
	resp, err := network.PerformReq(req, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	log.Println("response Body:", string(respBody))

	var ok ResponseOk
	json.Unmarshal(respBody, &ok)
	if len(ok.Data) == 0 {
		return "", fmt.Errorf(string(respBody))
	}
	return ok.Data[0], nil
}
