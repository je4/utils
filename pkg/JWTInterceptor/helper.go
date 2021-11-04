package JWTInterceptor

import (
	"fmt"
	"github.com/pkg/errors"
	"hash"
	"net/url"
	"sort"
	"strings"
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
			result += fmt.Sprintf("%s:%s;", k, val)
		}
	}
	return
}
