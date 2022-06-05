package captcha

import (
	"dollwipe/network"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"
)

// Timeout in seconds between two GET requests to RuCaptcha servers.
const cTimeout = 3

type RuCaptcha struct {
	Status  int32
	Request string
}

func RuCaptchaSolver(img []byte, key string) (string, error) {
	answer, err := ruCaptchaPost(img, key)
	if err != nil {
		return "", fmt.Errorf("RuCaptcha: ошибка: %v", err)
	}
	if answer.Status != 1 {
		return "", fmt.Errorf("RuCaptcha: не удалось отправить капчу: code = %d; %s",
			answer.Status, answer.Request)
	}
	var (
		id     = answer.Request
		errors = 0 // If request failed with error, then it will be resend "limit" times.
		limit  = 3
		failed = func(msg string, args ...interface{}) (string, error) {
			return "", fmt.Errorf(msg, args...)
		}
	)
	for {
		get, err := ruCaptchaGet(id, key)
		if err != nil {
			errors++
			if errors >= limit {
				return failed("RuCaptcha: GET запрос провалился %d раз(а), err = %v", limit, err)
			}
			continue
		}
		if get.Status == 1 {
			return get.Request, nil
		}
		switch get.Request {
		case "CAPCHA_NOT_READY":
			break
		case "ERROR_CAPTCHA_UNSOLVABLE":
			return failed("не удалось решить капчу, слишком мудрёная.")
		case "ERROR_WRONG_USER_KEY":
			return failed("неверный формат API ключа.")
		case "ERROR_KEY_DOES_NOT_EXIST":
			return failed("API ключ не существует.")
		// other errros described here: rucaptcha.com/api-rucaptcha#res_errors
		default:
			return failed("ошибка решения капчи: %s", get.Request)
		}
		time.Sleep(time.Second * time.Duration(cTimeout))
	}
}

func ruCaptchaPost(img []byte, key string) (RuCaptcha, error) {
	var (
		answer RuCaptcha
		params = map[string]string{
			"method": "post",
			"key":    key,
			"json":   "1",
		}
		files = network.FileForm{"file", map[string][]byte{
			"file": img,
		}}
		link = "http://rucaptcha.com/in.php"
	)
	req, err := network.NewPostMultipartRequest(link, params, &files)
	if err != nil {
		return answer, fmt.Errorf("не удалось составить POST запрос. err = %v", err)
	}
	resp, err := network.PerformReq(req, nil)
	if err != nil {
		return answer, fmt.Errorf("не удалось отправить POST запрос. err = %v", err)
	}
	defer resp.Body.Close()
	cont, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return answer, fmt.Errorf("не удалось прочесть ответ. err = %v", err)
	}
	json.Unmarshal(cont, &answer)
	return answer, nil
}

func ruCaptchaGet(id, key string) (RuCaptcha, error) {
	var (
		answer RuCaptcha
		link   = "http://rucaptcha.com/res.php?key=" + key +
			"&action=get&json=1&id=" + id
	)
	cont, err := network.SendGet(link)
	if err != nil {
		return answer, err
	}
	json.Unmarshal(cont, &answer)
	return answer, nil
}
