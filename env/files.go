// files.go: helper utils for working with input configs, etc.

package env

import (
	"dollwipe/network"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

// Read file and split it's content by pattern.
func SplitFileContent(dir, pattern string) ([]string, error) {
	cont, err := ioutil.ReadFile(dir)
	if err != nil {
		return nil, err
	}
	return strings.Split(string(cont), pattern), nil
}

// Load to memory all the media files (.png, .jpg, etc) that file-path contains.
// 2 * 10^7 bytes is the size limit for single file.
func GetMedia(dir string) ([]File, error) {
	cont, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	if !strings.HasSuffix(dir, "/") {
		dir += "/"
	}
	var (
		files  []File
		failed = 0
		pred   = func(name string) bool {
			name = strings.ToLower(name)
			return strings.HasSuffix(name, ".jpg") ||
				strings.HasSuffix(name, ".png") ||
				strings.HasSuffix(name, ".jpeg") ||
				strings.HasSuffix(name, ".mp4") ||
				strings.HasSuffix(name, ".webm") ||
				strings.HasSuffix(name, ".gif")
		}
	)
	for _, file := range cont {
		if pred(file.Name()) {
			fname := dir + file.Name()
			cont, err := ioutil.ReadFile(fname)
			if err != nil {
				failed++
				continue
			}
			if len(cont) > 2e7 { // 20MB is the limit.
				FilesLogger.Logf("%s: размер файла превышает допустимый.", fname)
				failed++
				continue
			}
			files = append(files, File{Name: fname, Content: cont})
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("%s: не нашла подходящие файлы (.png, .mp4, etc.)", dir)
	}
	FilesLogger.Logf("%d/%d файлов инициализировано.", len(files), len(files)+failed)
	return files, nil
}

// Get all captions separated by double blank line.
func GetCaptions(dir string) ([]string, error) {
	return SplitFileContent(dir, "\n\n")
}

// Get all valid-formated proxies from directory.
func GetProxies(dir string, sessions int) ([]network.Proxy, error) {
	result := make([]network.Proxy, 0)
	proxies, err := SplitFileContent(dir, "\n")
	if err != nil {
		return result, fmt.Errorf("не смогла прочесть файл с проксями: err = %v", err)
	}
	for _, addr := range proxies {
		proxy, err := getProxy(addr)
		if err != nil {
			ProxiesLogger.Logf("%s: %v", addr, err)
			continue
		}
		for i := 0; i < sessions; i++ {
			result = append(result, proxy)
			proxy.SessionId++
		}
	}
	if len(result) == 0 {
		return result, fmt.Errorf("не смогла найти ни одной валидной прокси.")
	}
	return result, nil
}
