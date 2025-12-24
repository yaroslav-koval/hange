package base64

import (
	"encoding/base64"

	"github.com/yaroslav-koval/hange/domain/crypt"
)

func NewBase64Decryptor() crypt.Decryptor {
	return &base64Decryptor{}
}

type base64Decryptor struct{}

func (b base64Decryptor) Decrypt(value []byte) ([]byte, error) {
	dst := make([]byte, base64.StdEncoding.DecodedLen(len(value)))

	n, err := base64.StdEncoding.Decode(dst, value)
	if err != nil {
		return nil, err
	}

	return dst[:n], nil
}
