package JWTInterceptor

import (
	"bytes"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/op/go-logging"
	"hash"
	"io"
	"net/http"
	"strings"
	"sync"
)

var allowedContentTypes = []string{
	"text/xml",
	"application/xml",
	"application/json",
}

func checkToken(tokenStr string, jwtKey string, jwtAlg []string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		talg := token.Method.Alg()
		algOK := false
		for _, a := range jwtAlg {
			if talg == a {
				algOK = true
				break
			}
		}
		if !algOK {
			return false, fmt.Errorf("unexpected signing method (allowed are %v): %v", jwtAlg, token.Header["alg"])
		}

		return []byte(jwtKey), nil
	})
	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}
	if !token.Valid {
		return nil, fmt.Errorf("Token not valid: %s", tokenStr)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("Cannot get claims from token %s", tokenStr)
	}
	return claims, nil
}

func JWTInterceptor(service, function string, level JWTInterceptorLevel, handler http.Handler, jwtKey string, jwtAlg []string, h hash.Hash, log *logging.Logger) http.Handler {
	var hashLock sync.Mutex
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error

		// check for allowed content-type
		contentType := r.Header.Get("Content-type")
		var allowed = false
		for _, val := range allowedContentTypes {
			if val == contentType {
				allowed = true
				break
			}
		}
		var bs []byte
		// read the full body
		body := r.Body
		if body != nil {
			bs, err = io.ReadAll(body)
			if err != nil {
				http.Error(w, fmt.Sprintf("JWTInterceptor: cannot read body: %v", err), http.StatusInternalServerError)
				return
			}
			body.Close()
		}
		if !allowed && len(bs) > 0 {
			http.Error(w, fmt.Sprintf("JWTInterceptor: invalid content-type for body: %s", contentType), http.StatusInternalServerError)
			return
		}

		// extract authorization bearer
		reqToken := r.Header.Get("Authorization")
		if reqToken != "" {
			splitToken := strings.Split(reqToken, "Bearer ")
			if len(splitToken) != 2 {
				http.Error(w, fmt.Sprintf("JWTInterceptor: no Bearer in Authorization header: %v", reqToken), http.StatusForbidden)
				return
			} else {
				reqToken = splitToken[1]
			}
		} else {
			reqToken = r.URL.Query().Get("token")
			if reqToken == "" {
				http.Error(w, fmt.Sprintf("JWTInterceptor: no token: %v", reqToken), http.StatusForbidden)
				return
			}
		}
		claims, err := checkToken(reqToken, jwtKey, jwtAlg)
		if err != nil {
			http.Error(w, fmt.Sprintf("JWTInterceptor: error in authorization token: %v", err), http.StatusForbidden)
			return
		}

		if service != claims["service"] {
			http.Error(w, fmt.Sprintf("JWTInterceptor: invalid service: %s != %s", service, claims["service"]), http.StatusForbidden)
			return
		}
		if function != claims["function"] {
			http.Error(w, fmt.Sprintf("JWTInterceptor: invalid function: %s != %s", function, claims["function"]), http.StatusForbidden)
			return
		}

		if level == Secure {
			hashLock.Lock()
			checksumBytes, err := buildHash(h, service, function, r.Method, checksumQueryValuesString(r.URL.Query()), bs)
			if err != nil {
				hashLock.Unlock()
				msg := fmt.Sprintf("JWTInterceptor: cannot write to checksum: %v", err)
				log.Errorf(msg)
				http.Error(w, msg, http.StatusInternalServerError)
				return
			}
			hashLock.Unlock()

			checksum := fmt.Sprintf("%x", checksumBytes)

			if checksum != claims["checksum"] {
				http.Error(w, fmt.Sprintf("JWTInterceptor: invalid checksum: %s != %s", checksum, claims["checksum"]), http.StatusForbidden)
				return
			}
		}

		// rewind body
		r.Body.Close()
		buf := bytes.NewBuffer(bs)
		r.Body = io.NopCloser(buf)

		//serve
		handler.ServeHTTP(w, r)
	})
}
