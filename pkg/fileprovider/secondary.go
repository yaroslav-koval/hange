package fileprovider

import "context"

type FileContentProvider interface {
	GetFileContent(context.Context, string) ([]byte, error)
}
