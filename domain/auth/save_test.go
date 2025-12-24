package auth

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	decryptor_mock "github.com/yaroslav-koval/hange/mocks/decryptor"
	encryptor_mock "github.com/yaroslav-koval/hange/mocks/encryptor"
	tokenfetcher_mock "github.com/yaroslav-koval/hange/mocks/tokenfetcher"
	tokenstorer_mock "github.com/yaroslav-koval/hange/mocks/tokenstorer"
)

func TestSaveToken(t *testing.T) {
	t.Parallel()

	storeErr := errors.New("store failed")
	encodedSecret := []byte("encoded-secret")
	trimmedToken := "secret-value"

	tests := []struct {
		name           string
		token          string
		setupStorer    func(storer *tokenstorer_mock.MockTokenStorer)
		setupEncryptor func(encryptor *encryptor_mock.MockEncryptor)
		wantErr        error
	}{
		{
			name:  "stores encrypted trimmed token",
			token: "secret-value\n",
			setupStorer: func(storer *tokenstorer_mock.MockTokenStorer) {
				storer.EXPECT().Store(string(encodedSecret)).Return(nil)
			},
			setupEncryptor: func(encryptor *encryptor_mock.MockEncryptor) {
				encryptor.EXPECT().
					Encrypt([]byte(trimmedToken)).
					Return(encodedSecret, nil)
			},
		},
		{
			name:  "returns error when store fails",
			token: "secret-value",
			setupStorer: func(storer *tokenstorer_mock.MockTokenStorer) {
				storer.EXPECT().Store(string(encodedSecret)).Return(storeErr)
			},
			setupEncryptor: func(encryptor *encryptor_mock.MockEncryptor) {
				encryptor.EXPECT().
					Encrypt([]byte(trimmedToken)).
					Return(encodedSecret, nil)
			},
			wantErr: storeErr,
		},
		{
			name:        "rejects empty token",
			token:       " \n\t",
			setupStorer: func(storer *tokenstorer_mock.MockTokenStorer) {},
			wantErr:     ErrEmptyToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockStorer := tokenstorer_mock.NewMockTokenStorer(t)
			mockEncryptor := encryptor_mock.NewMockEncryptor(t)
			mockDecryptor := decryptor_mock.NewMockDecryptor(t)

			if tt.setupStorer != nil {
				tt.setupStorer(mockStorer)
			}
			if tt.setupEncryptor != nil {
				tt.setupEncryptor(mockEncryptor)
			}

			auth := NewAuth(
				mockStorer,
				tokenfetcher_mock.NewMockTokenFetcher(t),
				mockEncryptor,
				mockDecryptor,
			)

			err := auth.SaveToken(tt.token)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			if errors.Is(tt.wantErr, ErrEmptyToken) {
				mockStorer.AssertNotCalled(t, "Store", mock.Anything)
				mockEncryptor.AssertNotCalled(t, "Encrypt", mock.Anything)
			}
		})
	}
}
