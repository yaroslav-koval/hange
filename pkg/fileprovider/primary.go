package fileprovider

import (
	"context"

	"github.com/yaroslav-koval/hange/pkg/entities"
)

type FileProvider interface {
	ReadFiles(context.Context, []string) (<-chan entities.File, <-chan error)
}

// TODO take values from env or config.
// readerWorkers := 3
//
//	if len(fileNames) < 10 {
//		readerWorkers = 1
//	}

type Config struct {
	Workers    int
	BufferSize int
}
