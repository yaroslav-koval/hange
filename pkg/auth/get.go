package auth

func (s *service) GetToken() (string, error) {
	return s.tokenFetcher.Fetch()
}
