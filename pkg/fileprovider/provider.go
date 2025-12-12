package fileprovider

import "context"

type FileProvider interface {
	ReadFiles(context.Context, []string) (<-chan File, <-chan error, error)
}

type File struct {
	FilePath string
	File     []byte
}
