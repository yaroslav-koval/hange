package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	tokenfetcher_mock "github.com/yaroslav-koval/hange/mocks/tokenfetcher"
	tokenstorer_mock "github.com/yaroslav-koval/hange/mocks/tokenstorer"
)

func TestSaveToken(t *testing.T) {
	t.Parallel()

	storeErr := errors.New("store failed")

	tests := []struct {
		name    string
		token   string
		setup   func(storer *tokenstorer_mock.MockTokenStorer)
		wantErr error
	}{
		{
			name:  "stores trimmed token",
			token: "secret-value\n",
			setup: func(storer *tokenstorer_mock.MockTokenStorer) {
				storer.EXPECT().Store("secret-value").Return(nil)
			},
		},
		{
			name:  "returns error when store fails",
			token: "secret-value",
			setup: func(storer *tokenstorer_mock.MockTokenStorer) {
				storer.EXPECT().Store("secret-value").Return(storeErr)
			},
			wantErr: storeErr,
		},
		{
			name:    "rejects empty token",
			token:   " \n\t",
			setup:   func(storer *tokenstorer_mock.MockTokenStorer) {},
			wantErr: ErrEmptyToken,
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStorer := tokenstorer_mock.NewMockTokenStorer(t)
			if tt.setup != nil {
				tt.setup(mockStorer)
			}

			auth := NewAuth(mockStorer, tokenfetcher_mock.NewMockTokenFetcher(t))

			err := auth.SaveToken(tt.token)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if tt.wantErr == ErrEmptyToken {
				mockStorer.AssertNotCalled(t, "Store", mock.Anything)
			}
		})
	}
}
