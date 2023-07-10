package grpc

import (
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/serviceconfig"
	"sync"
)

// derived from
// Timmermans, Jille (2022) grpc-multi-resolver source code (Version 1.10) [Source code]. https://github.com/Jille/grpc-multi-resolver

type partialClientConnGroup struct {
	cc    resolver.ClientConn
	parts []*partialClientConn
}

func (pccg *partialClientConnGroup) updateState() {
	s := resolver.State{}
	pccg.parts[0].mtx.Lock()
	s.ServiceConfig = pccg.parts[0].state.ServiceConfig
	s.Attributes = pccg.parts[0].state.Attributes
	pccg.parts[0].mtx.Unlock()
	for _, p := range pccg.parts {
		p.mtx.Lock()
		s.Addresses = append(s.Addresses, p.state.Addresses...)
		p.mtx.Unlock()
	}
	pccg.cc.UpdateState(s)
}

type partialClientConn struct {
	parent *partialClientConnGroup

	mtx   sync.Mutex
	state resolver.State
}

// UpdateState updates the state of the ClientConn appropriately.
func (cc *partialClientConn) UpdateState(s resolver.State) error {
	cc.mtx.Lock()
	cc.state = s
	cc.mtx.Unlock()
	cc.parent.updateState()
	return nil
}

// ReportError notifies the ClientConn that the Resolver encountered an
// error.  The ClientConn will notify the load balancer and begin calling
// ResolveNow on the Resolver with exponential backoff.
func (cc *partialClientConn) ReportError(err error) {
	cc.parent.cc.ReportError(err)
}

// NewAddress is called by resolver to notify ClientConn a new list
// of resolved addresses.
// The address list should be the complete list of resolved addresses.
//
// Deprecated: Use UpdateState instead.
func (cc *partialClientConn) NewAddress(addresses []resolver.Address) {
	cc.mtx.Lock()
	cc.state.Addresses = addresses
	cc.mtx.Unlock()
	cc.parent.updateState()
}

// NewServiceConfig is called by resolver to notify ClientConn a new
// service config. The service config should be provided as a json string.
//
// Deprecated: Use UpdateState instead.
func (cc *partialClientConn) NewServiceConfig(serviceConfig string) {
	cc.mtx.Lock()
	cc.state.ServiceConfig = cc.ParseServiceConfig(serviceConfig)
	cc.mtx.Unlock()
	cc.parent.updateState()
}

// ParseServiceConfig parses the provided service config and returns an
// object that provides the parsed config.
func (cc *partialClientConn) ParseServiceConfig(serviceConfigJSON string) *serviceconfig.ParseResult {
	return cc.parent.cc.ParseServiceConfig(serviceConfigJSON)
}
