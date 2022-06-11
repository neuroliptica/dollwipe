package main

import (
	"dollwipe/engine"
	"dollwipe/env"
	"dollwipe/network"
	"log"
	"math"
	"os"
	"time"
)

// Isolated goroutine for logging.
func logger(lenv *env.Env) {
	for v := range lenv.Logger {
		log.Println(v)
	}
}

// Isolated goroutine for statistics.
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
	go logger(lenv)
	go counter(lenv)

	// Init posts. Also if we do not use proxy, "localhost" will be count as a proxy in proxy map.
	// Despite this, it will never be set as a normal proxy.
	// So all the request will be performed through our own ip.
	engine.Posts = make(map[string]*engine.Post, 0)
	if !lenv.UseProxy {
		localhost := network.Proxy{"localhost", nil, "", ""}
		lenv.Proxies = append(lenv.Proxies, localhost) // So mod won't be zero
	}
	for _, proxy := range lenv.Proxies {
		post := engine.InitPost(lenv, proxy)
		engine.Posts[proxy.Addr] = post
	}
	if lenv.UseProxy {
		log.Printf("проксей инициализировано - %d.", len(lenv.Proxies))
	}

	for i := uint64(0); i < lenv.Iters; i++ {
		var (
			alive = make([]string, 0)
			used  = uint64(0)
			need  = uint64(math.Min(float64(len(engine.Posts)), float64(lenv.Threads)))
			shift = need * i                  // If threads < proxies, then we choose proxy to launch with shift.
			mod   = uint64(len(lenv.Proxies)) // Cycle array index.
		)
		for j := shift % mod; used < need; j = (j + 1) % mod {
			proxy := lenv.Proxies[j].Addr
			if _, ok := engine.Posts[proxy]; ok {
				alive = append(alive, proxy)
				used++
			}
		}
		log.Printf("итерация %d; постов будет отправлено - %d; перерыв - %d сек.", i+1, used, lenv.Timeout)
		for j := uint64(0); j < used; j++ {
			go engine.RunPost(engine.Posts[alive[j]])
		}
		for {
			if uint64(lenv.PostsOk+lenv.PostsFailed) == used {
				break
			}
			time.Sleep(time.Second * 2)
		}
		log.Printf("Успешно отправлено - %d; всего отправлено - %d.", lenv.PostsOk, lenv.PostsOk+lenv.PostsFailed)
		lenv.PostsOk, lenv.PostsFailed = 0, 0
		if len(engine.Posts) == 0 {
			log.Println("все проксичи умерли, помянем.")
			os.Exit(0)
		}
		if i+1 != lenv.Iters {
			time.Sleep(time.Second * time.Duration(lenv.Timeout))
		}
	}
}
