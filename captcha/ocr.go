// ocr.go: build and send request to the local ocr instance.

package captcha

import (
	"dollwipe/network"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
)

// Json schema for request. Body contains images in base64.
type RequestBody struct {
	Data []string `json:"data"`
}

// Json schema for response. Data contains solver result.
type ResponseOk struct {
	Data          []string
	Is_generating bool
	Duration      float64
}

// Local OCR instance url.
const NeuralUrl = "http://127.0.0.1:7860/api/predict"

// Solve captcha using ocr instance. Second argument is not used.
func NeuralSolver(img []byte, key string) (string, error) {
	body := RequestBody{
		Data: []string{"data:image/png;base64," + base64.StdEncoding.EncodeToString(img)},
	}
	payload, err := json.Marshal(body)
	if err != nil {
		return "", err
	}
	cont, err := network.SendPost(
		NeuralUrl,
		payload,
		map[string]string{
			"Content-Type": "application/json",
			"Connection":   "keep-alive",
		},
		nil)
	if err != nil {
		return "", err
	}

	log.Println("response Body:", string(cont))

	var ok ResponseOk
	json.Unmarshal(cont, &ok)
	if len(ok.Data) == 0 {
		return "", fmt.Errorf(string(cont))
	}
	return ok.Data[0], nil
}
