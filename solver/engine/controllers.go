package engine

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type Payload interface {
}

func JsonController[P Payload](w http.ResponseWriter, payload P) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}

func ReadJsonBody[P Payload](r *http.Request, payload *P) error {
	cont, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}
	json.Unmarshal(cont, payload)
	return nil
}

// /post
func PostController(w http.ResponseWriter, r *http.Request) {
	var req PostRequest
	err := ReadJsonBody(r, &req)
	if req.Base64 == "" || err != nil {
		JsonController(w, PostResponse{
			Error: MakeErrorBody("INVALID_BODY", err),
		})
		return
	}
	JsonController(w, PostModel(req))
}

// /get
func GetController(w http.ResponseWriter, r *http.Request) {
	JsonController(w, GetModel())
}

// /solve
func SolveController(w http.ResponseWriter, r *http.Request) {
	var req SolveRequest
	err := ReadJsonBody(r, &req)
	if req.Hash == "" || err != nil {
		JsonController(w, SolveResponse{
			Error: MakeErrorBody("INVALID_BODY", err),
		})
		return
	}
	JsonController(w, SolveModel(req))
}

// /check
func CheckController(w http.ResponseWriter, r *http.Request) {
	var req CheckRequest
	err := ReadJsonBody(r, &req)
	if req.Hash == "" || err != nil {
		JsonController(w, SolveResponse{
			Error: MakeErrorBody("INVALID_BODY", err),
		})
		return
	}
	JsonController(w, CheckModel(req))
}
