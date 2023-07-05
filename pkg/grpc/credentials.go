package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
)

func NewTLSCredentials(caCertPem []byte) (credentials.TransportCredentials, error) {
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(caCertPem) {
		return nil, errors.New("failed to add server CA's certificate")
	}

	// Create the credentials and return it
	config := &tls.Config{
		RootCAs: certPool,
	}

	return credentials.NewTLS(config), nil
}
