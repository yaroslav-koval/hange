package auth

func (s *service) GetToken() (string, error) {
	token, err := s.tokenFetcher.Fetch()
	if err != nil {
		return "", err
	}

	decrToken, err := s.decryptor.Decrypt([]byte(token))
	if err != nil {
		return "", err
	}

	return string(decrToken), nil
}
