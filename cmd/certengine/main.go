package main

import (
	"crypto/x509/pkix"
	"emperror.dev/emperror"
	"flag"
	"fmt"
	"github.com/je4/utils/v2/data/ca"
	"github.com/je4/utils/v2/pkg/cert"
	"github.com/pkg/errors"
	"net"
	"os"
	"strings"
	"time"
)

var ipsFlag = flag.String("ip", "", "semicolon separated list of ip addresses")
var dnsFlag = flag.String("dns", "", "semicolon separated list of dns names")
var name = flag.String("name", "", "filename prefix for cert an key")

func main() {
	flag.Parse()

	if *name == "" {
		emperror.Panic(errors.New("please provide a filename prefix (-name)"))
	}

	dns := strings.Split(*dnsFlag, ";")
	ipStrings := strings.Split(*ipsFlag, ";")

	certName := pkix.Name{
		Organization:  []string{"University Basel"},
		Country:       []string{"CH"},
		Province:      []string{"BS"},
		Locality:      []string{"Basel"},
		StreetAddress: []string{"Schönbeinstrasse 18-20"},
		PostalCode:    []string{"4056"},
	}
	dnsNames := append([]string{"localhost"}, dns...)
	ips := []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	for _, ipStr := range ipStrings {
		if ipStr == "" {
			continue
		}
		ipStr = strings.TrimSpace(ipStr)
		ip := net.ParseIP(ipStr)
		if ip == nil {
			emperror.Panic(errors.Errorf("invalid ip '%s'", ipStr))
		}
		ips = append(ips, ip)
	}

	srvPem, srvKey, err := cert.CreateServer(
		ca.CACert,
		ca.CAKey,
		certName,
		ips,
		dnsNames,
		time.Hour*24*364*10)
	if err != nil {
		panic(err)
	}

	certPath := *name + ".cert.pem"
	if err := os.WriteFile(certPath, srvPem, 0700); err != nil {
		panic(err)
	}
	keyPath := *name + ".key.pem"
	if err := os.WriteFile(keyPath, srvKey, 0700); err != nil {
		panic(err)
	}

	fmt.Printf("Certificate: %s\nPrivate Key: %s\n", certPath, keyPath)
}
