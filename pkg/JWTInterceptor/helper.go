package JWTInterceptor

import (
	"fmt"
	"net/url"
	"sort"
)

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
