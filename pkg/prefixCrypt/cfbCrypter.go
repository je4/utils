package prefixCrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"emperror.dev/errors"
)

func NewCFBCryptor(key []byte, iv []byte) (*cfbCrypt, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &cfbCrypt{
		block: block,
		iv:    iv,
	}, nil
}

type cfbCrypt struct {
	iv    []byte
	block cipher.Block
}

func (c *cfbCrypt) Encrypt(src []byte) ([]byte, error) {
	cfb := cipher.NewCFBEncrypter(c.block, c.iv)
	dst := make([]byte, len(src))
	cfb.XORKeyStream(dst, src)
	return dst, nil
}

func (c *cfbCrypt) Decrypt(src []byte) ([]byte, error) {
	cfb := cipher.NewCFBDecrypter(c.block, c.iv)
	dst := make([]byte, len(src))
	cfb.XORKeyStream(dst, src)
	return dst, nil
}

var _ Encrypter = (*cfbCrypt)(nil)
var _ Decrypter = (*cfbCrypt)(nil)
