package grpc

import (
	"context"
	"github.com/je4/utils/v2/pkg/cert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func UnaryCertInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		p, ok := peer.FromContext(ctx)
		if ok {
			tlsInfo := p.AuthInfo.(credentials.TLSInfo)
			subject := tlsInfo.State.VerifiedChains[0][0].Subject
			for _, s := range subject.ToRDNSequence() {
				for _, i := range s {
					if v, ok := i.Value.(string); ok {
						if i.Type.Equal(cert.OIDASN1UnstructuredName) {
							if v == info.FullMethod {
								return handler(ctx, req)
							}
						}
					}
				}
			}
			return nil, status.Errorf(codes.Unauthenticated, "method '%s' not in subject extraNames", info.FullMethod)
		}
		return nil, status.Errorf(codes.Unauthenticated, "no client certificate")
	}
}

func StreamCertInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		p, ok := peer.FromContext(ss.Context())
		if ok {
			tlsInfo := p.AuthInfo.(credentials.TLSInfo)
			subject := tlsInfo.State.VerifiedChains[0][0].Subject
			for _, s := range subject.ToRDNSequence() {
				for _, i := range s {
					if v, ok := i.Value.(string); ok {
						if i.Type.Equal(cert.OIDASN1UnstructuredName) {
							if v == info.FullMethod {
								return handler(srv, ss)
							}
						}
					}
				}
			}
			return status.Errorf(codes.Unauthenticated, "method '%s' not in subject extraNames", info.FullMethod)
		}
		return status.Errorf(codes.Unauthenticated, "no client certificate")
	}
}
