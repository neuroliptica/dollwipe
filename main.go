package main

import (
	"dollwipe/cache"
	"dollwipe/engine"
	"dollwipe/env"
	"dollwipe/network"
	"fmt"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/neuroliptica/logger"
)

const (
	POST_FAILED = iota
	POST_OK
	logo = `
     _       _ _          _            
    | |     | | |        (_)           
  __| | ___ | | |_      ___ _ __   ___ 
 / _' |/ _ \| | \ \ /\ / / | '_ \ / _ \
| (_| | (_) | | |\ V  V /| | |_) |  __/
 \__,_|\___/|_|_| \_/\_/ |_| .__/ \___|
                           | |         
                           |_|         
	`
)

var (
	InitLogger    = logger.MakeLogger("init").BindToDefault()
	InfoLogger    = logger.MakeLogger("info").BindToDefault()
	CookiesLogger = logger.MakeLogger("cookies").BindToDefault()
)

// Save alive proxies before exit.
func CacheAliveProxies(posts map[network.Proxy]*engine.Post) {
	filename := "alive_proxies.cache"
	err := cache.PostsCache(posts).CachePack(filename)
	if err != nil {
		cache.CacheLogger.Logf("не смогла сохранить живые прокси: %v", err)
		return
	}
	cache.CacheLogger.Logf("сохранила живые прокси => %s", filename)
}

// Check proxies for validness.
func CheckProxies(proxies []network.Proxy) []network.Proxy {
	var Checker sync.WaitGroup
	for i := range proxies {
		Checker.Add(1)
		go func(proxy *network.Proxy) {
			proxy.CheckAlive(time.Second * 60)
			Checker.Done()
		}(&proxies[i])
	}
	Checker.Wait()
	validProxies := make([]network.Proxy, 0)
	for i := range proxies {
		if proxies[i].Alive {
			validProxies = append(validProxies, proxies[i])
		}
	}
	return validProxies
}

func main() {
	fmt.Println(logo)
	lenv, err := env.ParseEnv()
	if err != nil {
		InitLogger.Logf("ошибка инициализации: %v", err)
		os.Exit(0)
	}

	// Statistics counter.
	postsUpdate := make(chan int)
	go func() {
		for ok := range lenv.Status {
			if ok {
				postsUpdate <- POST_OK
			} else {
				postsUpdate <- POST_FAILED
			}
		}
	}()

	// Init posts. Also if we do not use proxy then "localhost"
	// will be count as a proxy in proxy map. Despite this, it
	// will never be set as a normal proxy. So all the requests
	// will be performed through our own ip.
	var (
		Posts      = make(map[network.Proxy]*engine.Post, 0)
		PostsMutex sync.Mutex // Between filter goroutine and main.
	)
	if !lenv.UseProxy {
		localhost := network.Proxy{
			Addr: "localhost",
		}
		lenv.Proxies = append(lenv.Proxies, localhost)
	}

	// Filter invalid proxies before initialization.
	InitLogger.Log("предварительная проверка всех проксей...")
	validProxies := CheckProxies(lenv.Proxies)
	if len(validProxies) == 0 {
		InitLogger.Log("ни одна прокся не прошла первичную проверку, ошибка.")
		os.Exit(0)
	}
	InitLogger.Logf("%d/%d проксей будут инициализированы.",
		len(validProxies), len(lenv.Proxies))

	// This part will spawn goroutine for every Post instance.
	// Then will wait until Posts initialization is not finished.
	initResponse := make(chan engine.InitPostResponse)

	var SingleInit sync.WaitGroup
	go func() {
		failed := 0
		for v := range initResponse {
			if v.Post() == nil {
				failed++
			} else {
				sort.Slice(v.Post().Cookies, func(i, j int) bool {
					return v.Post().Cookies[i].Name < v.Post().Cookies[j].Name
				})
				Posts[v.Proxy] = v.Post()
			}
			CookiesLogger.Logf("OK: %3d; FAIL: %3d", len(Posts), failed)
			SingleInit.Done()

			if failed+len(Posts) == len(validProxies) {
				return
			}
		}
	}()

	// Init partially; InitAtOnce is corresponding to -I flag value.
	for i := 0; i < len(validProxies); i += int(lenv.InitAtOnce) {
		for j := 0; j < int(lenv.InitAtOnce) && i+j < len(validProxies); j++ {
			SingleInit.Add(1)
			go engine.InitPost(lenv, validProxies[i+j], initResponse)
		}
		SingleInit.Wait()
	}

	if len(Posts) == 0 {
		InitLogger.Log("ошибка, не удалось инициализировать ни одной прокси.")
		os.Exit(0)
	}
	if lenv.UseProxy {
		env.ProxiesLogger.Logf("проксей инициализировано - %d.", len(Posts))
	}

	// Thread safe bad proxies filter.
	go func() {
		for proxy := range lenv.Filter {
			PostsMutex.Lock()
			delete(Posts, proxy)
			PostsMutex.Unlock()
		}
	}()

	for i := uint64(0); i < lenv.Iters; i++ {
		PostsMutex.Lock() // Block for filter until init done.
		var (
			alive = make([]network.Proxy, 0)
			used  = uint64(0)
			need  = uint64(math.Min(float64(len(Posts)), float64(lenv.Threads)))
			shift = need * i                  // If threads < proxies, then we choose proxy to launch with shift.
			mod   = uint64(len(validProxies)) // Cycle array index.
		)
		for j := shift % mod; used < need; j = (j + 1) % mod {
			proxy := validProxies[j]
			if _, ok := Posts[proxy]; ok {
				alive = append(alive, proxy)
				used++
			}
		}
		InfoLogger.Logf("итерация %d; постов будет отправлено - %d; перерыв - %d сек",
			i+1, used, lenv.Timeout)

		for j := uint64(0); j < used; j++ {
			go engine.RunPost(Posts[alive[j]])
		}
		//
		PostsMutex.Unlock()

		postsOk, postsFail := 0, 0
		for uint64(postsOk+postsFail) != used {
			update := <-postsUpdate
			if update == POST_OK {
				postsOk++
			} else {
				postsFail++
			}
		}
		InfoLogger.Logf("успешно отправлено - %d; всего отправлено - %d.",
			postsOk, postsOk+postsFail)

		PostsMutex.Lock() // Wait until filter is done.
		if len(Posts) == 0 {
			InfoLogger.Log("все проксичи умерли, помянем.")
			os.Exit(0)
		}
		PostsMutex.Unlock()

		if i+1 != lenv.Iters {
			time.Sleep(time.Second * time.Duration(lenv.Timeout))
		}
	}
	CacheAliveProxies(Posts)
}
