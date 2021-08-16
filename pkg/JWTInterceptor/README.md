JWTInterceptor
==============

The `JWTInterceptor` package provides support for
automatic JWT Authorization of HTTP REST.

There's an implementation for HTTP-server (Handler) 
and -client (Roundtrip)

server example
--------------

	hello := func() http.Handler {
		return http.HandlerFunc(
            func(w http.ResponseWriter, 
                req *http.Request) { 
                    fmt.Fprintf(w, "hello\n") 
                })
	}

	mux := http.NewServeMux()
	mux.Handle("/test/sub", 
        JWTInterceptor.JWTInterceptor(hello(), 
            "/test", 
            "secret", 
            []string{"HS256", "HS384", "HS512"}, 
            sha512.New()))

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

client example
--------------

	tr, err := JWTInterceptor.NewJWTTransport(nil,
		sha512.New(),
		"/test",
		"secret",
		"HS512", 
        30*time.Second)
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