package factory

import (
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/auth/tokenfetch"
	"github.com/yaroslav-koval/hange/pkg/auth/tokenstore"
)

type App struct {
	Auth auth.Auth
}

func NewCLIApp() App {
	return App{
		auth.NewAuth(
			tokenstore.NewConfigTokenStorer(),
			tokenfetch.NewConfigTokenFetcher(),
		),
	}
}
