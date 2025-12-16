package fileprovider

import "context"

type FileNamesProvider interface {
	// GetAllFileNames accepts both file paths and directory paths. Returns only file paths after a recursive read.
	GetAllFileNames(context.Context, []string) ([]string, error)
}

type FileContentProvider interface {
	GetFileContent(context.Context, string) ([]byte, error)
}
