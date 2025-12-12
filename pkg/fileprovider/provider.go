package fileprovider

import (
	"context"

	"github.com/yaroslav-koval/hange/pkg/entities"
)

type FileProvider interface {
	ReadFiles(context.Context, []string) (<-chan entities.File, <-chan error, error)
}
