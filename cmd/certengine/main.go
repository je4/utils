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
var days = flag.Int("days", 365, "certificate lifetime")
var server = flag.Bool("server", false, "create server certificate")
var client = flag.Bool("client", false, "create client certificate")
var serviceName = flag.String("service", "", "name of service (client cert only)")
var targetServices = flag.String("target", "", "name of target services (client cert only)")
var caFlag = flag.Bool("ca", false, "create certificate authority")

func main() {
	var err error
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
	if len(*serviceName) > 0 {
		certName.Names = []pkix.AttributeTypeAndValue{}
		for _, sn := range strings.Split(*serviceName, ",") {
			sn = strings.TrimSpace(sn)
			certName.Names = append(certName.Names, cert.NewASN1UnstructuredName(sn))
		}
	}
	if len(*targetServices) > 0 {
		certName.Names = []pkix.AttributeTypeAndValue{}
		for _, tn := range strings.Split(*targetServices, ",") {
			tn = strings.TrimSpace(tn)
			certName.ExtraNames = append(certName.ExtraNames, cert.NewASN1UnstructuredName(tn))
		}
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

	var certPem []byte
	var certKey []byte

	var certType string
	if *server {
		certType = "server"
	}
	if *client {
		certType = "client"
	}
	if *caFlag {
		certType = "ca"
	}

	switch certType {
	case "server":
		certPem, certKey, err = cert.CreateServer(
			ca.CACert,
			ca.CAKey,
			certName,
			ips,
			dnsNames,
			time.Hour*24*time.Duration(*days))
	case "client":
		certPem, certKey, err = cert.CreateClient(
			ca.CACert,
			ca.CAKey,
			certName,
			time.Hour*24*time.Duration(*days))
	case "ca":
		certPem, certKey, err = cert.CreateCA(
			certName,
			time.Hour*24*time.Duration(*days))
	default:
		emperror.Panic(errors.New("please use -server or -client"))
	}
	if err != nil {
		panic(err)
	}

	certPath := *name + ".cert.pem"
	if err := os.WriteFile(certPath, certPem, 0700); err != nil {
		panic(err)
	}
	keyPath := *name + ".key.pem"
	if err := os.WriteFile(keyPath, certKey, 0700); err != nil {
		panic(err)
	}

	fmt.Printf("Certificate: %s\nPrivate Key: %s\n", certPath, keyPath)
}
