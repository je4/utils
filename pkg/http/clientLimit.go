package http

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/pkg/errors"
	"golang.org/x/net/http2"
	"net/http"
	"sync"
	"time"
)

func NewLimitedClient(concurrency int64, requestTimeout, tokenTimeout time.Duration, caCert []byte) *LimitedClient {

	caCertPool := x509.NewCertPool()
	if caCert != nil {
		caCertPool.AppendCertsFromPEM(caCert)
	}

	// Create TLS configuration with the certificate of the server
	tlsConfig := &tls.Config{
		RootCAs: caCertPool,
	}
	tr := &http2.Transport{
		TLSClientConfig: tlsConfig,
		AllowHTTP:       false,
		//		DialTLS: func(netw, addr string, cfg *tls.Config) (net.Conn, error) {
		//			return net.Dial(netw, addr)
		//		},
	}
	var lc = &LimitedClient{
		max:    concurrency,
		tokens: make(chan int64, concurrency),
		client: &http.Client{
			Transport: tr,
			Timeout:   requestTimeout,
		},
		tokenTimeout: tokenTimeout,
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
	sync.Mutex
	max          int64
	client       *http.Client
	tokens       chan int64
	tokenTimeout time.Duration
	//	counter      int64
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
			time.Sleep(time.Second)
			//			lc.client.CloseIdleConnections()
			resp, err = lc.client.Do(req)
			if err != nil {
				lc.tokens <- token
				return nil, errors.WithStack(err)
			}
		}
		return &Response{
			Response: resp,
			tokens:   lc.tokens,
			token:    token,
		}, nil
	case <-time.After(lc.tokenTimeout):
		return nil, errors.Errorf("timeout waiting for request token")
	}
}
