package JWTInterceptor

import (
	"bytes"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"hash"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type RoundTripper struct {
	http.RoundTripper
	hash             hash.Hash
	hashLock         sync.Mutex
	ignorePrefix     string
	jwtKey           string
	jwtSigningMethod jwt.SigningMethod
	lifetime         time.Duration
}

type Claims struct {
	Checksum string `json:"checksum"`
	jwt.StandardClaims
}

func NewJWTTransport(originalTransport http.RoundTripper,
	h hash.Hash,
	ignorePrefix string,
	jwtKey string, jwtAlg string,
	lifetime time.Duration) (http.RoundTripper, error) {
	if originalTransport == nil {
		originalTransport = http.DefaultTransport
	}
	tr := &RoundTripper{
		RoundTripper:     originalTransport,
		hash:             h,
		ignorePrefix:     ignorePrefix,
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

func (t *RoundTripper) buildHash(req *http.Request, data []byte) ([]byte, error) {
	t.hashLock.Lock()
	defer t.hashLock.Unlock()
	// create hash value
	t.hash.Reset()
	// checksum from method + path + url query params + body
	if _, err := t.hash.Write([]byte(strings.ToUpper(req.Method))); err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("cannot write method to checksum"))
	}
	if _, err := t.hash.Write([]byte(strings.TrimPrefix(req.URL.Path, t.ignorePrefix))); err != nil {
		return nil, errors.Wrapf(err, "cannot write rawquery to checksum")
	}
	if _, err := t.hash.Write([]byte(checksumQueryValuesString(req.URL.Query()))); err != nil {
		return nil, errors.Wrapf(err, "cannot write rawquery to checksum")
	}
	if data == nil {
		data = []byte{}
	}
	if _, err := t.hash.Write(data); err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("cannot write body to checksum"))
	}
	hashbytes := t.hash.Sum(nil)

	return hashbytes, nil
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

	hashBytes, err := t.buildHash(req, bodyBytes)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot build hash")
	}

	// create token
	claims := &Claims{
		Checksum: fmt.Sprintf("%x", hashBytes),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(t.lifetime).Unix(),
			Issuer:    "JWTInterceptor",
		},
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
