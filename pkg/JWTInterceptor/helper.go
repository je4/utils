package JWTInterceptor

import (
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
	"hash"
	"net/url"
	"sort"
	"strings"
	"time"
)

func buildHash(h hash.Hash, service, function, method, query string, data []byte) ([]byte, error) {
	// create hash value
	h.Reset()
	// checksum from service + method + url query params + body
	if _, err := h.Write([]byte(service)); err != nil {
		return nil, errors.Wrapf(err, "cannot write rawquery to checksum")
	}
	if _, err := h.Write([]byte(function)); err != nil {
		return nil, errors.Wrapf(err, "cannot write rawquery to checksum")
	}
	if _, err := h.Write([]byte(strings.ToUpper(method))); err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("cannot write method to checksum"))
	}
	if _, err := h.Write([]byte(query)); err != nil {
		return nil, errors.Wrapf(err, "cannot write rawquery to checksum")
	}
	if data == nil {
		data = []byte{}
	}
	if _, err := h.Write(data); err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("cannot write body to checksum"))
	}
	hashbytes := h.Sum(nil)

	return hashbytes, nil

}

func checksumQueryValuesString(values url.Values) (result string) {
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	// sort keys
	sort.Strings(keys)
	for _, k := range keys {
		vals := values[k]
		// sort values
		sort.Strings(vals)
		for _, val := range vals {
			// do not use token value...
			if k == "token" {
				continue
			}
			result += fmt.Sprintf("%s:%s;", k, val)
		}
	}
	return
}

func createToken(
	url *url.URL,
	method,
	service,
	function string,
	data []byte,
	lifetime time.Duration,
	h hash.Hash,
	jwtKey string,
	jwtSigningMethod jwt.SigningMethod,
	level JWTInterceptorLevel) (string, error) {
	var err error

	// create token
	claims := &Claims{
		Service:  service,
		Function: function,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(lifetime).Unix(),
			Issuer:    "JWTInterceptor",
		},
	}
	if level == Secure {
		hashBytes, err := buildHash(h, service, function, method, checksumQueryValuesString(url.Query()), data)
		if err != nil {
			return "", errors.Wrapf(err, "error building hash")
		}
		claims.Checksum = fmt.Sprintf("%x", hashBytes)
	}
	token := jwt.NewWithClaims(jwtSigningMethod, claims)
	ss, err := token.SignedString([]byte(jwtKey))
	if err != nil {
		return "", errors.Wrapf(err, "cannot sign token")
	}
	return ss, nil
}
