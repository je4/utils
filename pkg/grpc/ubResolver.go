package grpc

import "google.golang.org/grpc/resolver"

// derived from
// Timmermans, Jille (2022) grpc-multi-resolver source code (Version 1.10) [Source code]. https://github.com/Jille/grpc-multi-resolver

type ubResolver struct {
	children []resolver.Resolver
}

// ResolveNow will be called by gRPC to try to resolve the target name
// again. It's just a hint, resolver can ignore this if it's not necessary.
//
// It could be called multiple times concurrently.
func (m ubResolver) ResolveNow(opts resolver.ResolveNowOptions) {
	for _, r := range m.children {
		r.ResolveNow(opts)
	}
}

// Close closes the resolver.
func (m ubResolver) Close() {
	for _, r := range m.children {
		r.Close()
	}
}
