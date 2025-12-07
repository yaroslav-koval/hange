package auth

type Auth interface {
	SaveToken(authToken string) error
	GetToken() (string, error)
}

func NewAuth(tokenStorer TokenStorer, tokenFetcher TokenFetcher) Auth {
	return &service{
		tokenStorer:  tokenStorer,
		tokenFetcher: tokenFetcher,
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
}
