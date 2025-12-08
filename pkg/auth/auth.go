package auth

import "github.com/yaroslav-koval/hange/pkg/crypt"

type Auth interface {
	SaveToken(authToken string) error
	GetToken() (string, error)
}

func NewAuth(
	tokenStorer TokenStorer,
	tokenFetcher TokenFetcher,
	encryptor crypt.Encryptor,
	decryptor crypt.Decryptor) Auth {
	return &service{
		tokenStorer:  tokenStorer,
		tokenFetcher: tokenFetcher,
		encryptor:    encryptor,
		decryptor:    decryptor,
	}
}

type TokenStorer interface {
	Store(token string) error
}

type TokenFetcher interface {
	Fetch() (string, error)
}

type service struct {
	tokenStorer  TokenStorer
	tokenFetcher TokenFetcher
	encryptor    crypt.Encryptor
	decryptor    crypt.Decryptor
}
