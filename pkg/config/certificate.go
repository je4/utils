package config

import (
	"crypto/x509"
	"emperror.dev/errors"
	"encoding/pem"
	"os"
	"strings"
)

type Certificate x509.Certificate

func (cp *Certificate) UnmarshalText(text []byte) error {
	pemString := strings.TrimSpace(string(text))
	if found := envRegexp.FindStringSubmatch(pemString); found != nil {
		pemString = os.Getenv(found[1])
		if pemString == "" {
			return errors.Errorf("environment variable %s is empty", found[1])
		}
	} else {
		if !strings.HasPrefix(pemString, "-----BEGIN CERTIFICATE-----") {
			fi, err := os.Stat(pemString)
			if err != nil {
				if os.IsNotExist(err) {
					return errors.Errorf("'%s' not a certificate", pemString)
				}
				return errors.Wrapf(err, "cannot stat file %s", pemString)
			} else {
				if fi.IsDir() {
					return errors.Errorf("file %s is a directory", pemString)
				}
				data, err := os.ReadFile(pemString)
				if err != nil {
					return errors.Wrapf(err, "cannot read file %s", pemString)
				}
				pemString = string(data)
			}
		}
	}
	for block, rest := pem.Decode([]byte(pemString)); block != nil; block, rest = pem.Decode(rest) {
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return errors.Wrap(err, "cannot parse certificate")
		}
		*cp = Certificate(*cert)
		return nil
	}
	return errors.New("no certificate found")
}
