package engine

type (
	SolveRequest struct {
		Hash  string `json:"hash"`
		Value string `json:"value"`
	}
	SolveResponse struct {
		Error ErrorBody `json:"error"`
	}
)

func SolveModel(req SolveRequest) SolveResponse {
	DataMutex.Lock()
	defer DataMutex.Unlock()
	if _, ok := Queued[req.Hash]; !ok {
		return SolveResponse{
			Error: MakeErrorBody("TIMEOUT", nil),
		}
	}
	delete(Queued, req.Hash)
	Solved[req.Hash] = req.Value

	return SolveResponse{}
}
