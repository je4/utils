package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/pkg/errors"
	"google.golang.org/grpc/credentials"
)

func NewClientTLSCredentials(caCertPem []byte) (credentials.TransportCredentials, error) {
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

func NewServerTLSCredentials(serverCertPem, serverKeyPem []byte) (credentials.TransportCredentials, error) {
	cert, err := tls.X509KeyPair(serverCertPem, serverKeyPem)
	if err != nil {
		return nil, errors.Wrap(err, "cannot create certificate from pem")
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}
