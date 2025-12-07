package factory

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appfactory_mock "github.com/yaroslav-koval/hange/mocks/appfactory"
	configurator_mock "github.com/yaroslav-koval/hange/mocks/configurator"
	tokenfetcher_mock "github.com/yaroslav-koval/hange/mocks/tokenfetcher"
	tokenstorer_mock "github.com/yaroslav-koval/hange/mocks/tokenstorer"
)

func TestBuildApp(t *testing.T) {
	t.Run("builds app with provided dependencies", func(t *testing.T) {
		t.Parallel()

		mockFactory := appfactory_mock.NewMockAppFactory(t)
		cfg := configurator_mock.NewMockConfigurator(t)
		storer := tokenstorer_mock.NewMockTokenStorer(t)
		fetcher := tokenfetcher_mock.NewMockTokenFetcher(t)

		mockFactory.EXPECT().CreateConfigurator().Return(cfg, nil)
		mockFactory.EXPECT().CreateTokenStorer(cfg).Return(storer, nil)
		mockFactory.EXPECT().CreateTokenFetcher(cfg).Return(fetcher, nil)

		storer.EXPECT().Store("secret").Return(nil)
		fetcher.EXPECT().Fetch().Return("secret", nil)

		app, err := BuildApp(mockFactory)
		require.NoError(t, err)

		assert.Same(t, cfg, app.Config)
		require.NoError(t, app.Auth.SaveToken("secret\n"))
		token, err := app.Auth.GetToken()
		require.NoError(t, err)
		assert.Equal(t, "secret", token)
	})

	t.Run("returns error when configurator fails", func(t *testing.T) {
		t.Parallel()

		mockFactory := appfactory_mock.NewMockAppFactory(t)
		expErr := errors.New("config err")
		mockFactory.EXPECT().CreateConfigurator().Return(nil, expErr)

		app, err := BuildApp(mockFactory)
		require.ErrorIs(t, err, expErr)
		assert.Equal(t, App{}, app)
	})

	t.Run("returns error when token storer fails", func(t *testing.T) {
		t.Parallel()

		mockFactory := appfactory_mock.NewMockAppFactory(t)
		cfg := configurator_mock.NewMockConfigurator(t)
		expErr := errors.New("storer err")

		mockFactory.EXPECT().CreateConfigurator().Return(cfg, nil)
		mockFactory.EXPECT().CreateTokenStorer(cfg).Return(nil, expErr)

		app, err := BuildApp(mockFactory)
		require.ErrorIs(t, err, expErr)
		assert.Equal(t, App{}, app)
	})

	t.Run("returns error when token fetcher fails", func(t *testing.T) {
		t.Parallel()

		mockFactory := appfactory_mock.NewMockAppFactory(t)
		cfg := configurator_mock.NewMockConfigurator(t)
		storer := tokenstorer_mock.NewMockTokenStorer(t)
		expErr := errors.New("fetcher err")

		mockFactory.EXPECT().CreateConfigurator().Return(cfg, nil)
		mockFactory.EXPECT().CreateTokenStorer(cfg).Return(storer, nil)
		mockFactory.EXPECT().CreateTokenFetcher(cfg).Return(nil, expErr)

		app, err := BuildApp(mockFactory)
		require.ErrorIs(t, err, expErr)
		assert.Equal(t, App{}, app)
	})
}
