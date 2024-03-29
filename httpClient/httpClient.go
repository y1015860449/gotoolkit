package httpClient

import "github.com/go-resty/resty/v2"

type Http interface {
	PostBytes(url string, headers map[string]string, data []byte) ([]byte, int, error)
	PostForm(url string, headers map[string]string, data map[string]string) ([]byte, int, error)
	// data `string`, `[]byte`, `struct`, `map`, `slice`, `io.Reader`
	PostJson(url string, headers map[string]string, data interface{}) ([]byte, int, error)
	Get(url string, headers map[string]string, data map[string]string) ([]byte, int, error)
}

type httpCli struct {
}

func NewHttp() Http {
	return &httpCli{}
}

func (h *httpCli) PostBytes(url string, headers map[string]string, data []byte) ([]byte, int, error) {
	setReq := func(req *resty.Request) {
		req.SetHeader("Accept", "application/octet-stream")
		req.SetHeader("Content-Type", "application/octet-stream")
		if len(data) > 0 {
			req.SetBody(data)
		}
	}
	return httpPostDo(url, headers, setReq)
}

func (h *httpCli) PostForm(url string, headers map[string]string, data map[string]string) ([]byte, int, error) {
	setReq := func(req *resty.Request) {
		req.SetHeader("Content-Type", "application/x-www-form-urlencoded")
		if len(data) > 0 {
			req.SetFormData(data)
		}
	}
	return httpPostDo(url, headers, setReq)
}

func (h *httpCli) PostJson(url string, headers map[string]string, data interface{}) ([]byte, int, error) {
	setReq := func(req *resty.Request) {
		req.SetHeader("Content-Type", "application/json")
		if data != nil {
			req.SetBody(data)
		}
	}
	return httpPostDo(url, headers, setReq)
}

func (h *httpCli) Get(url string, headers map[string]string, data map[string]string) ([]byte, int, error) {
	setReq := func(req *resty.Request) {
		if len(data) > 0 {
			req.SetQueryParams(data)
		}
	}
	return httpGetDo(url, headers, setReq)
}

func httpPostDo(url string, headers map[string]string, setReq func(req *resty.Request)) ([]byte, int, error) {
	client := resty.New()
	req := client.R()
	for k, v := range headers {
		req.SetHeader(k, v)
	}
	setReq(req)
	resp, err := req.Post(url)
	if err != nil {
		return nil, 0, err
	}
	return resp.Body(), resp.StatusCode(), nil
}

func httpGetDo(url string, headers map[string]string, setReq func(req *resty.Request)) ([]byte, int, error) {
	client := resty.New()
	req := client.R()
	for k, v := range headers {
		req.SetHeader(k, v)
	}
	setReq(req)
	resp, err := req.Get(url)
	if err != nil {
		return nil, 0, err
	}
	return resp.Body(), resp.StatusCode(), nil
}
