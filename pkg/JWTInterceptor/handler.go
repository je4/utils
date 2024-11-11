package JWTInterceptor

import (
	"bytes"
	"emperror.dev/errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/je4/utils/v2/pkg/zLogger"
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

func jwtInterceptor(r *http.Request, hashLock sync.Mutex, service, function string, level JWTInterceptorLevel, jwtKey string, jwtAlg []string, h hash.Hash, adminBearer string) (int, error) {
	var err error

	// extract authorization bearer
	reqToken := r.Header.Get("Authorization")
	if reqToken != "" {
		splitToken := strings.Split(reqToken, "Bearer ")
		if len(splitToken) != 2 {
			return http.StatusForbidden, errors.New(fmt.Sprintf("JWTInterceptor: no Bearer in Authorization header: %v", reqToken))
		} else {
			reqToken = splitToken[1]
		}
		if adminBearer != "" && reqToken == adminBearer {
			return http.StatusOK, nil
		}
	} else {
		reqToken = r.URL.Query().Get("token")
		if reqToken == "" {
			return http.StatusForbidden, errors.New(fmt.Sprintf("JWTInterceptor: no token: %v", reqToken))
		}
	}

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
			return http.StatusInternalServerError, errors.Wrap(err, "JWTInterceptor: cannot read body")
		}
		body.Close()
	}
	if !allowed && len(bs) > 0 {
		return http.StatusInternalServerError, errors.New(fmt.Sprintf("JWTInterceptor: invalid content-type for body: %s", contentType))
	}

	claims, err := checkToken(reqToken, jwtKey, jwtAlg)
	if err != nil {
		return http.StatusForbidden, errors.Wrap(err, "JWTInterceptor: error in authorization token")
	}

	if service != claims["service"] {
		return http.StatusForbidden, errors.New(fmt.Sprintf("JWTInterceptor: invalid service: %s != %s", service, claims["service"]))
	}
	if function != claims["function"] {
		return http.StatusForbidden, errors.New(fmt.Sprintf("JWTInterceptor: invalid function: %s != %s", function, claims["function"]))
	}

	if level == Secure {
		hashLock.Lock()
		checksumBytes, err := buildHash(h, service, function, r.Method, checksumQueryValuesString(r.URL.Query()), bs)
		if err != nil {
			hashLock.Unlock()
			return http.StatusInternalServerError, errors.Wrap(err, "JWTInterceptor: cannot write to checksum")
		}
		hashLock.Unlock()

		checksum := fmt.Sprintf("%x", checksumBytes)

		if checksum != claims["checksum"] {
			return http.StatusForbidden, errors.New(fmt.Sprintf("JWTInterceptor: invalid checksum: %s != %s", checksum, claims["checksum"]))
		}
	}

	// rewind body
	r.Body.Close()
	buf := bytes.NewBuffer(bs)
	r.Body = io.NopCloser(buf)
	return http.StatusOK, nil
}

func JWTInterceptorGIN(service, function string, level JWTInterceptorLevel, jwtKey string, jwtAlg []string, h hash.Hash, adminBearer string, log zLogger.ZLogger) gin.HandlerFunc {
	var hashLock sync.Mutex
	return gin.HandlerFunc(func(g *gin.Context) {
		status, err := jwtInterceptor(g.Request, hashLock, service, function, level, jwtKey, jwtAlg, h, adminBearer)
		if err != nil {
			log.Error().Err(err).Msg("JWTInterceptor: error")
			g.AbortWithError(status, err)
			return
		}
		g.Next()
	})
}

func JWTInterceptor(service, function string, level JWTInterceptorLevel, next http.Handler, jwtKey string, jwtAlg []string, h hash.Hash, adminBearer string, log zLogger.ZLogger) http.Handler {
	var hashLock sync.Mutex
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status, err := jwtInterceptor(r, hashLock, service, function, level, jwtKey, jwtAlg, h, adminBearer)
		if err != nil {
			log.Error().Err(err).Msg("JWTInterceptor: error")
			http.Error(w, err.Error(), status)
			return
		}
		next.ServeHTTP(w, r)
	})
}
