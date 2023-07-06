package grpc

import (
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"net"
	"sync"
)

type RegisterService func(srv *grpc.Server)

func Startup(addr string, token string, serverCertPEM, serverKeyPEM []byte, opts []grpc.ServerOption, register RegisterService, wg *sync.WaitGroup) (ShutdownService, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot listen to tcp %s", addr)
	}

	if opts == nil {
		opts = []grpc.ServerOption{}
	}
	if token != "" {
		opts = append(opts,
			grpc.UnaryInterceptor(JWTUnaryTokenInterceptor(token)),
			grpc.StreamInterceptor(JWTStreamTokenInterceptor(token)),
		)
	}
	if serverCertPEM != nil {
		creds, err := NewServerTLSCredentials(serverCertPEM, serverKeyPEM)
		if err != nil {
			return nil, errors.Wrap(err, "cannot create credentials")
		}
		opts = append(opts, grpc.Creds(creds))
	}
	grpcServer := grpc.NewServer(opts...)
	register(grpcServer)

	go func() {
		defer wg.Done()
		fmt.Printf("starting grpc server at %s\n", addr)
		if err := grpcServer.Serve(listener); err != nil {
			fmt.Printf("error starting grpc server at %s: %v\n", addr, err)
		}
		fmt.Println("service ended")
	}()
	return grpcServer, nil
}
