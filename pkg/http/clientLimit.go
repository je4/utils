package http

import (
	"github.com/pkg/errors"
	"net/http"
	"time"
)

func NewLimitedClient(concurrency int64, timeout time.Duration) *LimitedClient {
	var lc = &LimitedClient{
		max:    concurrency,
		tokens: make(chan int64, concurrency),
		client: &http.Client{
			Timeout: timeout,
		},
	}
	for i := int64(0); i < lc.max; i++ {
		lc.tokens <- i
	}
	return lc
}

type Response struct {
	*http.Response
	tokens chan int64
	token  int64
}

func (resp *Response) Close() error {
	defer func() { resp.tokens <- resp.token }()
	return resp.Body.Close()
}

type LimitedClient struct {
	max     int64
	client  *http.Client
	timeout time.Duration
	tokens  chan int64
}

func (lc *LimitedClient) Get(urlStr string) (*Response, error) {
	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	resp, err := lc.Do(req)
	return resp, errors.WithStack(err)
}

func (lc *LimitedClient) Do(req *http.Request) (*Response, error) {
	select {
	case token := <-lc.tokens:
		resp, err := lc.client.Do(req)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		return &Response{
			Response: resp,
			tokens:   lc.tokens,
			token:    token,
		}, nil
	case <-time.After(lc.timeout):
		return nil, errors.Errorf("timeout waiting for request token")
	}
}
