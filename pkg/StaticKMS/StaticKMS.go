package StaticKMS

import (
	"emperror.dev/errors"
	"github.com/tink-crypto/tink-go/v2/core/registry"
	"github.com/tink-crypto/tink-go/v2/tink"
	"strings"
)

const staticPrefix = "static://"

func NewClient(credential string) (registry.KMSClient, error) {
	client := &staticClient{
		credential: credential,
	}
	return client, nil
}

type staticClient struct {
	credential string
}

func (k staticClient) Supported(keyURI string) bool {
	return strings.HasPrefix(keyURI, staticPrefix)
}

func (k staticClient) GetAEAD(keyURI string) (tink.AEAD, error) {
	if !k.Supported(keyURI) {
		return nil, errors.Errorf("unsupported keyURI '%s'", keyURI)
	}

	return newStaticAEAD(k.credential), nil
}

var _ registry.KMSClient = (*staticClient)(nil)
