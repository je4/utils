package JWTInterceptor

import (
	"bytes"
	"crypto/sha512"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	hello := func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) { fmt.Fprintf(w, "hello\n") })
	}

	mux := http.NewServeMux()
	mux.Handle("/test/sub", JWTInterceptor(hello(), "/test", "secret", []string{"HS256", "HS384", "HS512"}, sha512.New()))
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
	tr, err := NewJWTTransport(nil,
		sha512.New(),
		"/test",
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

func TestHandlerGetParam(t *testing.T) {
	tr, err := NewJWTTransport(nil,
		sha512.New(),
		"/test",
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

func TestHandlerPostParam(t *testing.T) {
	tr, err := NewJWTTransport(nil,
		sha512.New(),
		"/test",
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
