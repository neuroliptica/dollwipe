package engine

type (
	CheckRequest struct {
		Hash string `json:"hash"`
	}
	CheckResponse struct {
		Value string    `json:"value"`
		Error ErrorBody `json:"error"`
	}
)

func CheckModel(req CheckRequest) CheckResponse {
	DataMutex.Lock()
	defer DataMutex.Unlock()
	if value, ok := Solved[req.Hash]; ok {
		return CheckResponse{
			Value: value,
		}
	}
	if _, ok := Queued[req.Hash]; ok {
		return CheckResponse{
			Error: MakeErrorBody("QUEUED", nil),
		}
	}
	if Unsolved.Has(req.Hash) {
		return CheckResponse{
			Error: MakeErrorBody("UNSOLVED", nil),
		}
	}
	return CheckResponse{
		Error: MakeErrorBody("FAILED", nil),
	}
}
