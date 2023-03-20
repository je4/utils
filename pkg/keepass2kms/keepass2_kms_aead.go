package keepass2kms

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"emperror.dev/errors"
	"encoding/hex"
	"fmt"
	"github.com/google/tink/go/tink"
	keepass "github.com/tobischo/gokeepasslib/v3"
	keepasswrappers "github.com/tobischo/gokeepasslib/v3/wrappers"
	"io"
	"strings"
)

func findRootGroup(grp *keepass.RootData, name string) *keepass.Group {
	for key, g := range grp.Groups {
		if g.Name == name {
			return &grp.Groups[key]
		}
	}
	return nil
}

func findSubGroup(grp *keepass.Group, name string) *keepass.Group {
	for key, g := range grp.Groups {
		if g.Name == name {
			return &grp.Groups[key]
		}
	}
	return nil
}

func findEntry(grp *keepass.Group, name string) *keepass.Entry {
	for key, e := range grp.Entries {
		if e.GetTitle() == name {
			return &grp.Entries[key]
		}
	}
	return nil
}

func mkValue(key string, value string) keepass.ValueData {
	return keepass.ValueData{Key: key, Value: keepass.V{Content: value}}
}

func mkProtectedValue(key string, value string) keepass.ValueData {
	return keepass.ValueData{
		Key:   key,
		Value: keepass.V{Content: value, Protected: keepasswrappers.NewBoolWrapper(true)},
	}
}

type keepass2AEAD struct {
	uri string
	db  *keepass.Database
}

func (k *keepass2AEAD) getKey() ([]byte, error) {
	parts := strings.Split(k.uri, "/")

	group := findRootGroup(k.db.Content.Root, parts[0])
	if group == nil {
		return nil, fmt.Errorf("key %s not found", k.uri)
	}

	for i := 1; i < len(parts)-1; i++ {
		nextGroup := findSubGroup(group, parts[i])
		if nextGroup == nil {
			return nil, fmt.Errorf("key %s not found", k.uri)
		}
		group = nextGroup
	}
	entry := findEntry(group, parts[len(parts)-1])
	if entry == nil {
		return nil, errors.Errorf("key %s not found", k.uri)
	}
	keyString := entry.GetPassword()
	if keyString == "" {
		return nil, errors.Errorf("no password in key %s", k.uri)
	}
	key, err := hex.DecodeString(keyString)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decode key %s", k.uri)
	}
	if len(key) != 32 {
		return nil, errors.Errorf("key %s has wrong length", k.uri)
	}
	return key, nil
}

func (k *keepass2AEAD) Encrypt(plaintext, associatedData []byte) ([]byte, error) {
	key, err := k.getKey()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get key for %s", k.uri)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create cipher for %s", k.uri)
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrapf(err, "failed to create nonce for %s", k.uri)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create GCM for %s", k.uri)
	}
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, associatedData)
	return append(nonce, ciphertext...), nil
}

func (k *keepass2AEAD) Decrypt(ciphertext, associatedData []byte) ([]byte, error) {
	if len(ciphertext) < 12 {
		return nil, errors.Errorf("ciphertext too short for %s", k.uri)
	}
	key, err := k.getKey()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get key for %s", k.uri)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create cipher for %s", k.uri)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create GCM for %s", k.uri)
	}
	plaintext, err := aesgcm.Open(nil, ciphertext[:12], ciphertext[12:], associatedData)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to decrypt ciphertext for %s", k.uri)
	}
	return plaintext, nil
}

func newKeepass2AEAD(uri string, db *keepass.Database) tink.AEAD {
	return &keepass2AEAD{
		uri: uri,
		db:  db,
	}
}

var _ tink.AEAD = (*keepass2AEAD)(nil)
