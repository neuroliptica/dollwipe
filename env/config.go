// config.go: parsing json cookie configs.

package env

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const (
	hkKey   = "07IFkGaPtry7mAxrnFzTs7kG"
	lifeKey = "cf_clearance"
)

type CookieLife struct {
	Key   string
	Value string `json:"cf_clearance"`
}

type CookieHk struct {
	Key   string
	Value string `json:"07IFkGaPtry7mAxrnFzTs7kG"`
}

func CookieParse(dir, domain string) ([]*http.Cookie, error) {
	cont, err := ioutil.ReadFile(dir)
	if err != nil {
		return nil, fmt.Errorf("не смогла прочесть файл: %v", err)
	}
	result := make([]*http.Cookie, 0)
	switch domain {
	case "hk":
		hk := CookieHk{
			Key: hkKey,
		}
		json.Unmarshal(cont, &hk)
		if hk.Value == "" {
			return nil, fmt.Errorf("config-hk.json: не смогла распарсить конфиг %s", dir)
		}
		result = append(result, &http.Cookie{Name: hk.Key, Value: hk.Value})
		return result, nil
	case "life":
		life := CookieLife{
			Key: lifeKey,
		}
		json.Unmarshal(cont, &life)
		if life.Value == "" {
			return nil, fmt.Errorf("config-life.json: не смогла распарсить конфиг %s", dir)
		}
		result = append(result, &http.Cookie{Name: life.Key, Value: life.Value})
		return result, nil
	default:
		return nil, fmt.Errorf("не смогла распознать домен: %s", domain)
	}
}
