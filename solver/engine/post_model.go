package engine

import "github.com/google/uuid"

type (
	PostRequest struct {
		Base64 string `json:"data"`
	}

	PostResponse struct {
		Hash  string    `json:"hash"`
		Error ErrorBody `json:"error"`
	}
)

func PostModel(req PostRequest) PostResponse {
	DataMutex.Lock()
	defer DataMutex.Unlock()

	hash := uuid.NewString()
	unsolved := Captcha{
		Hash:   hash,
		Base64: req.Base64,
	}
	Unsolved.Add(unsolved)
	return PostResponse{
		Hash: hash,
	}
}
