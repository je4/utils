package JWTInterceptor

import (
	"bytes"
	"crypto/sha512"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/op/go-logging"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	logger, _ := logging.GetLogger("test")
	hello := func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { fmt.Fprintf(w, "hello\n") })
	}

	mux := http.NewServeMux()
	mux.Handle("/test/sub", JWTInterceptor("test", "func", Secure, hello(), "secret", []string{"HS256", "HS384", "HS512"}, sha512.New(), logger))
	srv := &http.Server{
		Handler: mux,
		Addr:    ":7788",
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("server died: %v", err)
		}
	}()
	defer srv.Close()

	code := m.Run()
	os.Exit(code)
}

func TestHandlerGetBase(t *testing.T) {
	tr, err := NewJWTTransport(
		"test",
		"func",
		Secure,
		nil,
		sha512.New(),
		"secret",
		"HS512", 30*time.Second)
	if err != nil {
		t.Fatalf("cannot create new transport: %v", err)
	}
	client := &http.Client{
		Transport: tr,
	}

	resp, err := client.Get("http://localhost:7788/test/sub")
	if err != nil {
		t.Fatalf("webserver not running: %v", err)
	}
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("cannot read result: %v", err)
	}
	resultStr := strings.TrimSpace(string(result))
	if resultStr != "hello" {
		t.Fatalf("result: %s != hello", resultStr)
	}
}

func TestHandlerGetHeader(t *testing.T) {
	tr, err := NewJWTTransport(
		"test",
		"func",
		Secure,
		nil,
		sha512.New(),
		"secret",
		"HS512", 30*time.Second)
	if err != nil {
		t.Fatalf("cannot create new transport: %v", err)
	}
	client := &http.Client{
		Transport: tr,
	}

	resp, err := client.Get("http://localhost:7788/test/sub?blah=bl%20ubb")
	if err != nil {
		t.Fatalf("webserver not running: %v", err)
	}
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("cannot read result: %v", err)
	}
	resultStr := strings.TrimSpace(string(result))
	if resultStr != "hello" {
		t.Fatalf("result: %s != hello", resultStr)
	}

}

func TestGetParam(t *testing.T) {
	u, _ := url.Parse("http://localhost:7788/test/sub?blah=bl%20ubb")
	token, err := createToken(u, "GET", "test", "func", nil, 30*time.Second, sha512.New(), "secret", jwt.SigningMethodHS384, Secure)
	if err != nil {
		t.Fatalf("cannot create token: %v", err)
	}
	urlStr := u.String() + "&token=" + token
	resp, err := http.Get(urlStr)
	if err != nil {
		t.Fatalf("webserver not running: %v", err)
	}
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("cannot read result: %v", err)
	}
	resultStr := strings.TrimSpace(string(result))
	if resultStr != "hello" {
		t.Fatalf("result: %s != hello", resultStr)
	}
}

func TestHandlerGetParam(t *testing.T) {
	tr, err := NewJWTTransport(
		"test",
		"func",
		Secure,
		nil,
		sha512.New(),
		"secret",
		"HS512",
		30*time.Second,
	)
	if err != nil {
		t.Fatalf("cannot create new transport: %v", err)
	}
	client := &http.Client{
		Transport: tr,
	}

	resp, err := client.Get("http://localhost:7788/test/sub?blah=bl%20ubb")
	if err != nil {
		t.Fatalf("webserver not running: %v", err)
	}
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("cannot read result: %v", err)
	}
	resultStr := strings.TrimSpace(string(result))
	if resultStr != "hello" {
		t.Fatalf("result: %s != hello", resultStr)
	}

}

func TestHandlerPostParam(t *testing.T) {
	tr, err := NewJWTTransport(
		"test",
		"func",
		Secure,
		nil,
		sha512.New(),
		"secret",
		"HS512", 30*time.Second)
	if err != nil {
		t.Fatalf("cannot create new transport: %v", err)
	}
	client := &http.Client{
		Transport: tr,
	}
	var jsonStr = []byte(`{"title":"Testing 123"}`)
	resp, err := client.Post("http://localhost:7788/test/sub?blah=bl%20ubb",
		"application/json",
		bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatalf("webserver not running: %v", err)
	}
	result, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("cannot read result: %v", err)
	}
	resultStr := strings.TrimSpace(string(result))
	if resultStr != "hello" {
		t.Fatalf("result: %s != hello", resultStr)
	}

}
