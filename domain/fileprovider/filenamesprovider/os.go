package filenamesprovider

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/yaroslav-koval/hange/domain/fileprovider"
	"github.com/yaroslav-koval/hange/domain/fileprovider/errmapper"
)

func NewOSFileNamesProvider(errMapper errmapper.FileErrorMapper) fileprovider.FileNamesProvider {
	return &osFileNamesProvider{
		errMapper: errMapper,
	}
}

type osFileNamesProvider struct {
	errMapper errmapper.FileErrorMapper
}

func (o *osFileNamesProvider) GetAllFileNames(ctx context.Context, paths []string) ([]string, error) {
	var fileNames []string

	for _, p := range paths {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		fileInfo, err := os.Stat(p)
		if err != nil {
			return nil, o.createErrFailedPath(err, p)
		}

		if !fileInfo.IsDir() {
			fileNames = append(fileNames, p)

			continue
		}

		files, err := o.readFilesInDir(ctx, p)
		if err != nil {
			return nil, err
		}

		fileNames = append(fileNames, files...)
	}

	return fileNames, nil
}

// readFilesInDir reads all the files from the dir or recursively reads all children directories and their files
func (o *osFileNamesProvider) readFilesInDir(ctx context.Context, dirPath string) ([]string, error) {
	dir, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, o.createErrFailedPath(err, dirPath)
	}

	var fileNames []string

	for _, entry := range dir {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		currPath := filepath.Join(dirPath, entry.Name())

		if !entry.IsDir() {
			fileNames = append(fileNames, currPath)
			continue
		}

		entries, err := o.readFilesInDir(ctx, currPath)
		if err != nil {
			return nil, err
		}
		fileNames = append(fileNames, entries...)
	}

	return fileNames, nil
}

func (o *osFileNamesProvider) createErrFailedPath(err error, path string) error {
	return fmt.Errorf("failed to process path %s: %w", path, o.errMapper.Map(err))
}
