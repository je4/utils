package cert

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"emperror.dev/errors"
	"encoding/pem"
	"math/big"
	"net"
	"time"
)

func CreateServer(caPEM, caPrivKeyPEM []byte, certName pkix.Name, ips []net.IP, dnsNames []string, duration time.Duration) ([]byte, []byte, error) {
	var err error
	caPrivKeyBlock, _ := pem.Decode(caPrivKeyPEM)
	var exStart = "-----BEGIN EC"
	var caPrivKey any
	if string(caPrivKeyPEM[0:len(exStart)]) == exStart {
		caPrivKey, err = x509.ParseECPrivateKey(caPrivKeyBlock.Bytes)
		if err != nil {
			return nil, nil, errors.Wrap(err, "cannot decode ca private key")
		}
	} else {
		caPrivKey, err = x509.ParsePKCS1PrivateKey(caPrivKeyBlock.Bytes)
		if err != nil {
			return nil, nil, errors.Wrap(err, "cannot decode ca private key")
		}
	}
	caBlock, _ := pem.Decode(caPEM)
	caCerts, err := x509.ParseCertificates(caBlock.Bytes)
	if err != nil {
		return nil, nil, errors.Wrap(err, "cannot parse ca pem")
	}
	if len(caCerts) == 0 {
		return nil, nil, errors.New("no ca certificate in pem")
	}

	if ips == nil {
		ips = []net.IP{}
	}
	if len(ips) == 0 {
		ips = append(ips, net.IPv4(127, 0, 0, 1), net.IPv6loopback)
	}
	if dnsNames == nil {
		dnsNames = []string{}
	}
	if len(dnsNames) == 0 {
		dnsNames = append(dnsNames, "localhost")
	}
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject:      certName,
		IPAddresses:  ips,
		DNSNames:     dnsNames,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(duration),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	//certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCerts[0], &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pkBytes, err := x509.MarshalECPrivateKey(certPrivKey)
	if err != nil {
		return nil, nil, err
	}
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: pkBytes,
	})

	return certPEM.Bytes(), certPrivKeyPEM.Bytes(), nil
}
