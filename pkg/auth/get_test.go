package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	decryptor_mock "github.com/yaroslav-koval/hange/mocks/decryptor"
	encryptor_mock "github.com/yaroslav-koval/hange/mocks/encryptor"
	tokenfetcher_mock "github.com/yaroslav-koval/hange/mocks/tokenfetcher"
	tokenstorer_mock "github.com/yaroslav-koval/hange/mocks/tokenstorer"
)

func TestGet(t *testing.T) {
	t.Parallel()

	mockFetcher := tokenfetcher_mock.NewMockTokenFetcher(t)
	mockDecryptor := decryptor_mock.NewMockDecryptor(t)

	auth := NewAuth(
		tokenstorer_mock.NewMockTokenStorer(t),
		mockFetcher,
		encryptor_mock.NewMockEncryptor(t),
		mockDecryptor,
	)

	mockFetcher.EXPECT().Fetch().Return("cipher-text", nil)
	mockDecryptor.EXPECT().Decrypt([]byte("cipher-text")).Return([]byte("secret-value"), nil)

	actual, err := auth.GetToken()
	require.NoError(t, err)
	assert.Equal(t, "secret-value", actual)
}
