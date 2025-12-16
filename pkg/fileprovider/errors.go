package fileprovider

import (
	"errors"
)

var (
	ErrNotExist   = errors.New("not exists")
	ErrPermission = errors.New("permission denied")
)
