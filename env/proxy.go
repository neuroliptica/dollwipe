// proxy.go: parser for proxies.
// Check validness.

package env

import (
	"dollwipe/network"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"
)

var protocols = map[string]bool{
	"http":   true,
	"https":  true,
	"socks4": true,
	"socks5": true,
}

// protocol://n.n.n.n:port
// protocol://login:pass@n.n.n.n:port

func getProtocol(addr string) (protocol, rest string, err error) {
	arr := strings.Split(addr, ":")
	if len(arr) < 3 || len(arr[1]) <= 2 {
		err = fmt.Errorf("неверный формат.")
		return
	}
	if ok := protocols[arr[0]]; !ok {
		err = fmt.Errorf("неподдерживаемый/неизвестный протокол.")
		return
	}
	protocol = arr[0]
	arr[1] = arr[1][2:] // cut down '//'
	rest = strings.Join(arr[1:], ":")
	return
}

// If proxy is not authorized, then login and pass will be set to "".
func getCredits(addr string) (login, pass, rest string, err error) {
	arr := strings.Split(addr, "@")
	if len(arr) == 1 {
		rest = arr[0]
		return
	}
	credits := strings.Split(arr[0], ":")
	if len(credits) != 2 {
		err = fmt.Errorf("login/pass: неверный формат.")
		return
	}
	return credits[0], credits[1], arr[1], nil
}

func getAddress(addr string) (ip, port string, err error) {
	arr := strings.Split(addr, ":")
	if len(arr) < 2 {
		err = fmt.Errorf("неверный формат.")
		return
	}
	ip, port = arr[0], arr[1]
	nums := strings.Split(ip, ".")
	if len(nums) != 4 {
		err = fmt.Errorf("неверный формат.")
		return
	}
	for _, num := range nums {
		if _, err = strconv.Atoi(num); err != nil {
			err = fmt.Errorf("адрес содержит нечисловые литералы.")
			return
		}
	}
	runePort := []rune(port)
	for i, c := range runePort {
		if !unicode.IsDigit(c) {
			runePort = runePort[:i]
		}
	}
	port = string(runePort)
	if len(port) == 0 {
		err = fmt.Errorf("порт содержит нечисловые литералы.")
		return
	}
	return
}

func getProxy(addr string) (proxy network.Proxy, err error) {
	// Validating format
	protocol, rest, err := getProtocol(addr)
	if err != nil {
		return
	}
	login, pass, rest, err := getCredits(rest)
	if err != nil {
		return
	}
	ip, port, err := getAddress(rest)
	if err != nil {
		return
	}
	//
	addr = protocol + "://" + ip + ":" + port
	u, err := url.Parse(addr)
	if err != nil {
		return
	}
	proxy = network.Proxy{
		Addr:       addr,
		AddrParsed: u,
		Login:      login,
		Pass:       pass,
	}
	return
}
