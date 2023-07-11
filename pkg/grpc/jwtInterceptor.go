package grpc

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var errMissingMetadata = status.Errorf(codes.InvalidArgument, "no incoming metadata in rpc context")
var errMissingMetadataToken = status.Errorf(codes.InvalidArgument, "no token found in rpc metadata context")
var errInvalidMetadataToken = status.Errorf(codes.InvalidArgument, "invalid token found in rpc metadata context")

func JWTUnaryTokenInterceptor(token string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, errMissingMetadata
		}
		tokens := md.Get("Token")
		if len(tokens) == 0 {
			return nil, errMissingMetadataToken
		}
		if tokens[0] != token {
			return nil, errInvalidMetadataToken
		}

		md.Delete("Token")

		ctx = metadata.NewIncomingContext(ctx, md)

		return handler(ctx, req)
	}
}

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (s *wrappedStream) Context() context.Context {
	return s.ctx
}

func JWTStreamTokenInterceptor(token string) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return errMissingMetadata
		}

		tokens := md.Get("Token")
		if len(tokens) == 0 {
			return errMissingMetadataToken
		}
		if tokens[0] != token {
			return errInvalidMetadataToken
		}

		md.Delete("Token")

		ctx := metadata.NewIncomingContext(ss.Context(), md)

		return handler(srv, &wrappedStream{ss, ctx})
	}
}
