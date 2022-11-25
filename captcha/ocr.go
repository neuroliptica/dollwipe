package captcha

import (
	"bytes"
	"dollwipe/network"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
)

type PostBody struct {
	Fn_index     int32  `json:"fn_index"`
	Data         string `json:"data"`
	Session_hash string `json:"session_hash"`
}

const url = "https://hf.space/embed/sneedium/dvatch_captcha_sneedium/api/predict/"

func makeId(length int) string {
	var (
		result string
		chars  = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"
	)
	charl := len(chars)
	for i := 0; i < length; i++ {
		result += string(chars[rand.Intn(charl)])
	}
	return result
}

func NeuralSolver(img []byte, key string) (string, error) {
	body := PostBody{
		Fn_index:     0,
		Data:         base64.StdEncoding.EncodeToString(img),
		Session_hash: makeId(12),
	}
	bodyBytes, err := json.Marshal(body)
	log.Println(string(bodyBytes))
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Sec-Fetch-Dest", "empty")
	req.Header.Add("Sec-Fetch-Mode", "cors")
	req.Header.Add("Sec-Fetch-Site", "same-origin")
	req.Header.Add("Sec-GPC", "1")
	req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/101.0.4951.61 Safari/537.36")

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
	return "", nil
}
