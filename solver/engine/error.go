package engine

type ErrorBody struct {
	Failed   bool   `json:"failed"`
	Reason   string `json:"msg"`
	Internal string `json:"internal"`
}

func MakeErrorBody(reason string, err error) ErrorBody {
	body := ErrorBody{
		Failed: true,
		Reason: reason,
	}
	if err != nil {
		body.Internal = err.Error()
	}
	return body
}
