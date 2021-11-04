package JWTInterceptor

import (
	"bytes"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"hash"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

type JWTInterceptorLevel int

const Simple JWTInterceptorLevel = 0
const Secure JWTInterceptorLevel = 1

type RoundTripper struct {
	http.RoundTripper
	service          string
	function         string
	level            JWTInterceptorLevel
	hash             hash.Hash
	hashLock         sync.Mutex
	jwtKey           string
	jwtSigningMethod jwt.SigningMethod
	lifetime         time.Duration
}

type Claims struct {
	Checksum string `json:"checksum"`
	Service  string `json:"service"`
	Function string `json:"function"`
	jwt.StandardClaims
}

func NewJWTTransport(
	service, function string,
	level JWTInterceptorLevel,
	originalTransport http.RoundTripper,
	h hash.Hash,
	jwtKey string, jwtAlg string,
	lifetime time.Duration) (http.RoundTripper, error) {
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}
	tr := &RoundTripper{
		RoundTripper:     originalTransport,
		service:          service,
		function:         function,
		level:            level,
		hash:             h,
		jwtKey:           jwtKey,
		jwtSigningMethod: nil,
		lifetime:         lifetime,
	}
	switch jwtAlg {
	case "HS256":
		tr.jwtSigningMethod = jwt.SigningMethodHS256
	case "HS384":
		tr.jwtSigningMethod = jwt.SigningMethodHS384
	case "HS512":
		tr.jwtSigningMethod = jwt.SigningMethodHS512
	default:
		return nil, errors.New(fmt.Sprintf("invalid jwt algorithm: %s", jwtAlg))
	}
	return tr, nil
}

func (t *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	var err error

	// get the whole body
	body := req.Body
	var bodyBytes []byte
	if body != nil {
		bodyBytes, err = ioutil.ReadAll(body)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot read request body")
		}
		body.Close()
	}

	// create token
	claims := &Claims{
		Service:  t.service,
		Function: t.function,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(t.lifetime).Unix(),
			Issuer:    "JWTInterceptor",
		},
	}
	if t.level == Secure {
		t.hashLock.Lock()
		hashBytes, err := buildHash(t.hash, t.service, t.function, req.Method, checksumQueryValuesString(req.URL.Query()), bodyBytes)
		if err != nil {
			t.hashLock.Unlock()
			return nil, errors.Wrapf(err, "error building hash")
		}
		t.hashLock.Unlock()
		claims.Checksum = fmt.Sprintf("%x", hashBytes)
	}
	token := jwt.NewWithClaims(t.jwtSigningMethod, claims)
	ss, err := token.SignedString([]byte(t.jwtKey))
	if err != nil {
		return nil, errors.Wrapf(err, "cannot sign token")
	}

	// set header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ss))

	// set new body
	req.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes))

	// let the default do the work
	return t.RoundTripper.RoundTrip(req)
}
