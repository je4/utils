package grpc

import (
	"context"
	"google.golang.org/grpc/credentials"
)

func NewBearerAuth(key string) *jwtAuth {
	return &jwtAuth{key: key}
}

type jwtAuth struct {
	key string
}

func (b *jwtAuth) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	// todo: create jwt for uris
	return map[string]string{"Token": b.key}, nil
}

func (b *jwtAuth) RequireTransportSecurity() bool {
	return false
}

var (
	_ credentials.PerRPCCredentials = (*jwtAuth)(nil)
)
