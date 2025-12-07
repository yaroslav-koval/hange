package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGet(t *testing.T) {
	t.Parallel()

	mockFetcher := NewMockTokenFetcher(t)

	auth := NewAuth(NewMockTokenStorer(t), mockFetcher)

	mockFetcher.EXPECT().Fetch().Return("secret-value", nil)

	actual, err := auth.GetToken()
	require.NoError(t, err)
	assert.Equal(t, "secret-value", actual)
}
