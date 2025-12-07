package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
	tokenfetcher_mock "github.com/yaroslav-koval/hange/mocks/tokenfetcher"
	tokenstorer_mock "github.com/yaroslav-koval/hange/mocks/tokenstorer"
)

func TestSaveToken(t *testing.T) {
	t.Parallel()

	mockStorer := tokenstorer_mock.NewMockTokenStorer(t)

	auth := NewAuth(mockStorer, tokenfetcher_mock.NewMockTokenFetcher(t))

	mockStorer.EXPECT().Store("secret-value").Return(nil)

	err := auth.SaveToken("secret-value\n")
	require.NoError(t, err)
}
