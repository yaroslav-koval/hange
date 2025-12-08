package base64

import (
	"encoding/base64"

	"github.com/yaroslav-koval/hange/pkg/crypt"
)

func NewBase64Encryptor() crypt.Encryptor {
	return &base64Encryptor{}
}

type base64Encryptor struct{}

func (b base64Encryptor) Encrypt(value []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.EncodedLen(len(value)))

	base64.StdEncoding.Encode(dst, value)

	return dst, nil
}
