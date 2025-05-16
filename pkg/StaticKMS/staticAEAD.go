package StaticKMS

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"emperror.dev/errors"
	"github.com/tink-crypto/tink-go/v2/tink"
	"io"
)

type staticAEAD struct {
	credential string
}

func (k *staticAEAD) getKey() ([]byte, error) {
	return []byte(k.credential), nil
}

func (k *staticAEAD) Encrypt(plaintext, associatedData []byte) ([]byte, error) {
	key, err := k.getKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get static key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create static cipher")
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.Wrap(err, "failed to create static nonce")
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create static GCM")
	}
	ciphertext := aesgcm.Seal(nil, nonce, plaintext, associatedData)
	return append(nonce, ciphertext...), nil
}

func (k *staticAEAD) Decrypt(ciphertext, associatedData []byte) ([]byte, error) {
	if len(ciphertext) < 12 {
		return nil, errors.Errorf("ciphertext too short")
	}
	key, err := k.getKey()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get key")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create cipher")
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GCM")
	}
	plaintext, err := aesgcm.Open(nil, ciphertext[:12], ciphertext[12:], associatedData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to decrypt ciphertext")
	}
	return plaintext, nil
}

func newStaticAEAD(credential string) tink.AEAD {
	return &staticAEAD{
		credential: credential,
	}
}

var _ tink.AEAD = (*staticAEAD)(nil)
