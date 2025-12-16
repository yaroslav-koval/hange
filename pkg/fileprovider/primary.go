package fileprovider

import (
	"context"

	"github.com/yaroslav-koval/hange/pkg/entities"
)

type FileProvider interface {
	ReadFiles(context.Context, Config, []string) (<-chan entities.File, <-chan error)
	FileNamesProvider
}

type Config struct {
	Workers    int
	BufferSize int
}
