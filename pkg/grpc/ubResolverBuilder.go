package grpc

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
	"sync"
)

const ubResolverScheme = "static"

func NewUBResolver(addrs map[string][]string) *ubResolverBuilder {
	r := &ubResolverBuilder{addrs: addrs}
	return r
}

type ubResolverBuilder struct {
	sync.Mutex
	addrs         map[string][]string
	logger        *logging.Logger
	clientConn    resolver.ClientConn
	resolverConn  *grpc.ClientConn
	serviceConfig *serviceconfig.ParseResult
}

func (r *ubResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.Lock()
	defer r.Unlock()
	r.clientConn = cc
	var dialOpts = []grpc.DialOption{}
	if opts.DialCreds != nil {
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(opts.DialCreds))
	}
	r.serviceConfig = r.clientConn.ParseServiceConfig(
		fmt.Sprintf(`{"loadBalancingConfig":[{"%s":{}}]}`, ubResolverScheme),
	)
	var res = &staticResolver{clientConn: r.clientConn}
	var ok bool
	var endpoint = target.URL.Path
	if endpoint == "" {
		endpoint = target.URL.Opaque
	}
	res.endpoints, ok = r.addrs[endpoint]
	if !ok {
		return nil, errors.Errorf("invalid endpoint '%s'", endpoint)
	}
	res.ResolveNow(resolver.ResolveNowOptions{})

	return res, nil
}

func (r *ubResolverBuilder) Scheme() string { return ubResolverScheme }

var _ resolver.Builder = (*ubResolverBuilder)(nil)
