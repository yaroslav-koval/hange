package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tokenfetcher_mock "github.com/yaroslav-koval/hange/mocks/tokenfetcher"
	tokenstorer_mock "github.com/yaroslav-koval/hange/mocks/tokenstorer"
)

func TestGet(t *testing.T) {
	t.Parallel()

	mockFetcher := tokenfetcher_mock.NewMockTokenFetcher(t)

	auth := NewAuth(tokenstorer_mock.NewMockTokenStorer(t), mockFetcher)

	mockFetcher.EXPECT().Fetch().Return("secret-value", nil)

	actual, err := auth.GetToken()
	require.NoError(t, err)
	assert.Equal(t, "secret-value", actual)
}
