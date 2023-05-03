package engine

import (
	"log"
	"net/http"
)

func Run(address string) {
	http.HandleFunc("/post", PostController)
	http.HandleFunc("/get", GetController)
	http.HandleFunc("/solve", SolveController)
	http.HandleFunc("/check", CheckController)

	log.Fatal(http.ListenAndServe(address, nil))
}
