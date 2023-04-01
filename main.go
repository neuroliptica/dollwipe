package main

import (
	"dollwipe/engine"
	"dollwipe/env"
	"dollwipe/network"
	"fmt"
	"log"
	"math"
	"os"
	"time"
)

func logger(messages <-chan string) {
	for msg := range messages {
		log.Println(msg)
	}
}

func filter(bad <-chan string, posts map[string]*engine.Post) {
	for proxy := range bad {
		delete(posts, proxy)
	}
}

func counter(lenv *env.Env) {
	// TODO: should use mutex to prevent shit increment btw.
	for v := range lenv.Status {
		if v {
			lenv.PostsOk++
		} else {
			lenv.PostsFailed++
		}
	}
}

func main() {
	log.SetFlags(log.Ltime)
	lenv, err := env.ParseEnv()
	if err != nil {
		log.Fatal(err)
	}
	go logger(lenv.Logger)
	go counter(lenv)

	// Init posts. Also if we do not use proxy, "localhost" will be count as a proxy in proxy map.
	// Despite this, it will never be set as a normal proxy.
	// So all the request will be performed through our own ip.
	var Posts = make(map[string]*engine.Post, 0)
	if !lenv.UseProxy {
		localhost := network.Proxy{"localhost", nil, "", ""}
		lenv.Proxies = append(lenv.Proxies, localhost) // So mod won't be zero
	}

	// This part will spawn goroutine for every Post instance.
	// Then 'll wait until Posts initialization will finish.
	initResponse := make(chan engine.InitPostResponse)
	initDone := make(chan bool)
	go func(resp <-chan engine.InitPostResponse, done chan<- bool) {
		failed := 0
		for v := range resp {
			if v.Post() == nil {
				failed++
			} else {
				Posts[v.Address()] = v.Post()
			}
			lenv.Logger <- fmt.Sprintf(
				"OK: %3d; FAIL: %3d", len(Posts), failed)
			if failed+len(Posts) == len(lenv.Proxies) {
				done <- true
				return
			}
		}
	}(initResponse, initDone)

	for _, proxy := range lenv.Proxies {
		go engine.InitPost(lenv, proxy, initResponse)
	}
	// Block until initialization is done.
	<-initDone

	if lenv.UseProxy {
		lenv.Logger <- fmt.Sprintf(
			"проксей инициализировано - %d.", len(lenv.Proxies))
	}
	go filter(lenv.Filter, Posts)

	for i := uint64(0); i < lenv.Iters; i++ {
		var (
			alive = make([]string, 0)
			used  = uint64(0)
			need  = uint64(math.Min(float64(len(Posts)), float64(lenv.Threads)))
			shift = need * i                  // If threads < proxies, then we choose proxy to launch with shift.
			mod   = uint64(len(lenv.Proxies)) // Cycle array index.
		)
		for j := shift % mod; used < need; j = (j + 1) % mod {
			proxy := lenv.Proxies[j].Addr
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
		for uint64(lenv.PostsOk+lenv.PostsFailed) != used {
			time.Sleep(time.Second * 2)
		}
		lenv.Logger <- fmt.Sprintf(
			"Успешно отправлено - %d; всего отправлено - %d.",
			lenv.PostsOk, lenv.PostsOk+lenv.PostsFailed)
		lenv.PostsOk, lenv.PostsFailed = 0, 0
		if len(Posts) == 0 {
			lenv.Logger <- fmt.Sprintf("все проксичи умерли, помянем.")
			os.Exit(0)
		}
		if i+1 != lenv.Iters {
			time.Sleep(time.Second * time.Duration(lenv.Timeout))
		}
	}
}
