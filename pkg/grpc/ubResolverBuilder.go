package grpc

import (
	"fmt"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
	"net/url"
	"sync"
)

const ubResolverScheme = "ubbasel"

type ResolverEntity struct {
	Addr  []string
	Token string
	Ca    []byte
}

type ResolverData map[string]ResolverEntity

func NewUBResolver(addrs ResolverData) *UBResolverBuilder {
	r := &UBResolverBuilder{addrs: addrs}
	return r
}

type UBResolverBuilder struct {
	sync.Mutex
	addrs  ResolverData
	logger *logging.Logger
}

func (r *UBResolverBuilder) Dial(name string) (*grpc.ClientConn, error) {
	res, ok := r.addrs[name]
	if !ok {
		return nil, errors.Errorf("client '%s' not found", name)
	}
	var opts = []grpc.DialOption{}
	if res.Token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(NewBearerAuth(res.Token)))
	}
	if len(res.Ca) > 0 {
		creds, err := NewClientTLSCredentials(res.Ca)
		if err != nil {
			return nil, errors.Wrap(err, "cannot read CA PEM")
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}
	addr := fmt.Sprintf("%s:%s", ubResolverScheme, name)
	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot dial '%s'", addr)
	}
	return conn, nil
}

func parseTarget(target string) (ret resolver.Target) {
	u, err := url.Parse(target)
	if err != nil {
		u.RawPath = target
	}
	ret.URL = *u
	return ret
}

func (r *UBResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	//	r.Lock()
	//	defer r.Unlock()
	pccg := &partialClientConnGroup{
		cc: cc,
	}
	var ubRes = &ubResolver{}
	var ok bool
	var endpoint = target.URL.Path
	if endpoint == "" {
		endpoint = target.URL.Opaque
	}
	rd, ok := r.addrs[endpoint]
	if !ok {
		return nil, errors.Errorf("invalid endpoint '%s'", endpoint)
	}
	for _, addr := range rd.Addr {
		parsedAddr := parseTarget(addr)
		resolverBuilder := resolver.Get(parsedAddr.URL.Scheme)
		if resolverBuilder == nil {
			parsedAddr = resolver.Target{
				URL: url.URL{
					Scheme: resolver.GetDefaultScheme(),
					Path:   addr,
				},
			}
			resolverBuilder = resolver.Get(parsedAddr.URL.Scheme)
			if resolverBuilder == nil {
				return nil, fmt.Errorf("no resolver for default scheme: %q", parsedAddr.URL.Scheme)
			}
		}
		pcc := &partialClientConn{parent: pccg}
		pccg.parts = append(pccg.parts, pcc)
		res, err := resolverBuilder.Build(parsedAddr, pcc, opts)
		if err != nil {
			ubRes.Close()
			return nil, err
		}
		ubRes.children = append(ubRes.children, res)
	}

	//	ubRes.ResolveNow(resolver.ResolveNowOptions{})

	return ubRes, nil
}

func (r *UBResolverBuilder) Scheme() string { return ubResolverScheme }

var _ resolver.Builder = (*UBResolverBuilder)(nil)
