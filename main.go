package main

import (
	"dollwipe/engine"
	"dollwipe/env"
	"dollwipe/network"
	"fmt"
	"log"
	"math"
	"os"
	"sort"
	"time"
)

type (
	SingleInitDone struct{}
	InitDone       struct{}
)

const (
	POST_FAILED = iota
	POST_OK
)

func main() {
	log.SetFlags(log.Ltime)
	lenv, err := env.ParseEnv()
	if err != nil {
		log.Fatal(err)
	}
	// Thread safe logging purpose goroutine. All future logging
	// should be done through the lenv.Logger channel.
	//
	// ex: lenv.Logger <- log_message
	go func() {
		for msg := range lenv.Logger {
			log.Println(msg)
		}
	}()

	// Thread safe statistics counter.
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
	var Posts = make(map[network.Proxy]*engine.Post, 0)
	if !lenv.UseProxy {
		localhost := network.Proxy{
			Addr: "localhost",
		}
		lenv.Proxies = append(lenv.Proxies, localhost)
	}

	// This part will spawn goroutine for every Post instance.
	// Then will wait until Posts initialization is not finished.
	initResponse := make(chan engine.InitPostResponse)
	initDone := make(chan InitDone)
	singleInitDone := make(chan SingleInitDone)

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
			lenv.Logger <- fmt.Sprintf(
				"OK: %3d; FAIL: %3d", len(Posts), failed)
			singleInitDone <- SingleInitDone{}

			if failed+len(Posts) == len(lenv.Proxies) {
				initDone <- InitDone{}
				return
			}
		}
	}()

	// Init partially; InitAtOnce is corresponding to -I flag value.
	for i := 0; i < len(lenv.Proxies); i += int(lenv.InitAtOnce) {
		launched := 0
		for j := 0; j < int(lenv.InitAtOnce) && i+j < len(lenv.Proxies); j++ {
			go engine.InitPost(lenv, lenv.Proxies[i+j], initResponse)
			launched++
		}
		done := 0
		for _ = range singleInitDone {
			done++
			if done == launched {
				break
			}
		}
	}
	// Block until initialization is done.
	<-initDone

	if len(Posts) == 0 {
		lenv.Logger <- "ошибка, не удалось инициализировать ни одной прокси."
		os.Exit(0)
	}
	if lenv.UseProxy {
		lenv.Logger <- fmt.Sprintf(
			"проксей инициализировано - %d.", len(lenv.Proxies))
	}

	// Thread safe bad proxies filter.
	go func() {
		for proxy := range lenv.Filter {
			delete(Posts, proxy)
		}
	}()

	for i := uint64(0); i < lenv.Iters; i++ {
		var (
			alive = make([]network.Proxy, 0)
			used  = uint64(0)
			need  = uint64(math.Min(float64(len(Posts)), float64(lenv.Threads)))
			shift = need * i                  // If threads < proxies, then we choose proxy to launch with shift.
			mod   = uint64(len(lenv.Proxies)) // Cycle array index.
		)
		for j := shift % mod; used < need; j = (j + 1) % mod {
			proxy := lenv.Proxies[j]
			if _, ok := Posts[proxy]; ok {
				alive = append(alive, proxy)
				used++
			}
		}
		lenv.Logger <- fmt.Sprintf(
			"итерация %d; постов будет отправлено - %d; перерыв - %d сек.",
			i+1, used, lenv.Timeout)
		for j := uint64(0); j < used; j++ {
			go engine.RunPost(Posts[alive[j]])
		}

		postsOk, postsFail := 0, 0
		for uint64(postsOk+postsFail) != used {
			update := <-postsUpdate
			if update == POST_OK {
				postsOk++
			} else {
				postsFail++
			}
		}
		lenv.Logger <- fmt.Sprintf(
			"Успешно отправлено - %d; всего отправлено - %d.",
			postsOk, postsOk+postsFail)
		if len(Posts) == 0 {
			lenv.Logger <- "все проксичи умерли, помянем."
			os.Exit(0)
		}
		if i+1 != lenv.Iters {
			time.Sleep(time.Second * time.Duration(lenv.Timeout))
		}
	}
}
