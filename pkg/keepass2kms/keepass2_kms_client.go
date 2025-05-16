package keepass2kms

import (
	"emperror.dev/errors"
	"github.com/tink-crypto/tink-go/v2/core/registry"
	"github.com/tink-crypto/tink-go/v2/tink"
	keepass "github.com/tobischo/gokeepasslib/v3"
	"os"
	"strings"
)

const keepass2Prefix = "keepass2://"

// NewClientWithCredentials returns a new Keepass 2 client.
// uriPrefix must have the following format: 'keepass2://[:path]'.
// keyPath is the path inside the kdbx file to the key.
// kdbxPassword is the password to the kdbx file.
// credentials is the password to be stored in the key.
func NewClientWithCredentials(kdbx string, name string, credentials string) (registry.KMSClient, error) {
	/*
		if !strings.HasPrefix(strings.ToLower(kdbx), keepass2Prefix) {
			return nil, fmt.Errorf("uriPrefix must start with %s", keepass2Prefix)
		}
		fp, err := os.Open(kdbx[len(keepass2Prefix):])
	*/

	fp, err := os.Open(kdbx)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open keepass2 database %s", kdbx)

	}
	db := keepass.NewDatabase()
	db.Credentials = keepass.NewPasswordCredentials(credentials)
	if err := keepass.NewDecoder(fp).Decode(db); err != nil {
		return nil, errors.Wrapf(err, "cannot decode keepass2 database %s", kdbx)
	}
	if err := db.UnlockProtectedEntries(); err != nil {
		return nil, errors.Wrapf(err, "cannot unlock keepass2 database %s", kdbx)
	}

	return NewClient(db, name)
}

func NewClient(db *keepass.Database, name string) (registry.KMSClient, error) {
	client := &keepass2Client{
		db:   db,
		name: name,
	}
	return client, nil
}

type keepass2Client struct {
	db *keepass.Database
	//keyPath string
	name string
}

func (k keepass2Client) Supported(keyURI string) bool {
	return strings.HasPrefix(keyURI, keepass2Prefix+k.name+"/")
}

func (k keepass2Client) GetAEAD(keyURI string) (tink.AEAD, error) {
	if !k.Supported(keyURI) {
		return nil, errors.Errorf("unsupported keyURI '%s'", keyURI)
	}

	uri := keyURI[len(keepass2Prefix)+len(k.name)+1:]
	return newKeepass2AEAD(uri, k.db), nil
}

var _ registry.KMSClient = (*keepass2Client)(nil)
