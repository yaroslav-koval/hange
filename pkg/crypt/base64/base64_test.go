package base64_test

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	cryptbase64 "github.com/yaroslav-koval/hange/pkg/crypt/base64"
)

func TestBase64EncryptorEncryptsValue(t *testing.T) {
	encryptor := cryptbase64.NewBase64Encryptor()
	input := []byte("hello world")

	encoded, err := encryptor.Encrypt(input)
	require.NoError(t, err)
	require.Equal(t, base64.StdEncoding.EncodeToString(input), string(encoded))
	require.Len(t, encoded, base64.StdEncoding.EncodedLen(len(input)))
}

func TestBase64DecryptorDecryptsValue(t *testing.T) {
	decryptor := cryptbase64.NewBase64Decryptor()
	encoded := base64.StdEncoding.EncodeToString([]byte("hello world"))

	decoded, err := decryptor.Decrypt([]byte(encoded))
	require.NoError(t, err)
	require.Equal(t, "hello world", string(decoded))
	require.Len(t, decoded, len("hello world"))
}

func TestBase64DecryptorFailsOnInvalidInput(t *testing.T) {
	decryptor := cryptbase64.NewBase64Decryptor()

	decoded, err := decryptor.Decrypt([]byte("!!!"))
	require.Error(t, err)
	require.True(t, errors.Is(err, base64.CorruptInputError(0)))
	require.Nil(t, decoded)
}
