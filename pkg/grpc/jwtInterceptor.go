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

func JWTInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, errMissingMetadata
	}

	tokens := md.Get("Token")
	if len(tokens) == 0 {
		return nil, errMissingMetadataToken
	}
	//	info.FullMethod

	md.Delete("Token")

	ctx = metadata.NewIncomingContext(ctx, md)

	return handler(ctx, req)
}
