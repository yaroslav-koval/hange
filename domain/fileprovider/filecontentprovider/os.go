package filecontentprovider

import (
	"context"
	"os"

	"github.com/yaroslav-koval/hange/domain/fileprovider"
	"github.com/yaroslav-koval/hange/domain/fileprovider/errmapper"
)

func NewOSFileContentProvider(errMapper errmapper.FileErrorMapper) fileprovider.FileContentProvider {
	return &osFileContentProvider{
		errMapper: errMapper,
	}
}

type osFileContentProvider struct {
	errMapper errmapper.FileErrorMapper
}

func (o *osFileContentProvider) GetFileContent(ctx context.Context, filePath string) ([]byte, error) {
	res, err := os.ReadFile(filePath)
	if err != nil {
		return nil, o.errMapper.Map(err)
	}

	return res, nil
}
