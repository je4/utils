package grpc

import (
	"fmt"
	"google.golang.org/grpc/resolver"
	"sync"
)

type staticResolver struct {
	sync.Mutex
	endpoints  []string
	clientConn resolver.ClientConn
	name       string
}

func (r *staticResolver) ResolveNow(options resolver.ResolveNowOptions) {
	r.Lock()
	defer r.Unlock()
	var addrs = []resolver.Address{}
	for i, addr := range r.endpoints {
		addrs = append(addrs, resolver.Address{
			Addr:       addr,
			ServerName: fmt.Sprintf("%s-%d", r.name, i+1),
		})
	}
	r.clientConn.UpdateState(resolver.State{
		Addresses: addrs,
	})
}

func (r *staticResolver) Close() {

}

var _ resolver.Resolver = (*staticResolver)(nil)
