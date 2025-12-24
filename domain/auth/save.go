package auth

import (
	"errors"
	"log/slog"
	"strings"
)

var ErrEmptyToken = errors.New("empty token")

func (s *service) SaveToken(authToken string) error {
	if authToken = strings.TrimSpace(authToken); authToken == "" {
		return ErrEmptyToken
	}

	encrypted, err := s.encryptor.Encrypt([]byte(authToken))
	if err != nil {
		return err
	}

	if err = s.tokenStorer.Store(string(encrypted)); err != nil {
		slog.Info("Failed to store token")

		return err
	}

	slog.Info("Authentication token is stored")

	return nil
}
