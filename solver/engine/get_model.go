package engine

type (
	GetResponse struct {
		Base64 string    `json:"data"`
		Hash   string    `json:"hash"`
		Error  ErrorBody `json"error"`
	}
)

func GetModel() GetResponse {
	DataMutex.Lock()
	defer DataMutex.Unlock()

	captcha, err := Unsolved.Pop()
	if err != nil {
		return GetResponse{
			Error: MakeErrorBody("EMPTY", err),
		}
	}
	if _, ok := Queued[captcha.Hash]; ok {
		return GetResponse{
			Error: MakeErrorBody("ALREADY_IN_QUEUE", nil),
		}
	}
	Queued[captcha.Hash] = struct{}{}
	return GetResponse{
		Base64: captcha.Base64,
		Hash:   captcha.Hash,
	}
}
