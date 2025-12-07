package auth

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveToken(t *testing.T) {
	t.Parallel()

	mockStorer := NewMockTokenStorer(t)

	auth := NewAuth(mockStorer, NewMockTokenFetcher(t))

	mockStorer.EXPECT().Store("secret-value").Return(nil)

	err := auth.SaveToken("secret-value\n")
	require.NoError(t, err)
}
