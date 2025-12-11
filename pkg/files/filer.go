package files

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

var ErrNotExists = os.ErrNotExist

type File struct {
	FilePath string
	File     os.File
}

type DirectoryReader interface {
	// ReadFiles reads files and directories recursively. Second argument accepts both file and directory paths.
	ReadFiles(context.Context, []string) (<-chan File, error)
}

func NewDirectoryReader() DirectoryReader {
	return &directoryReader{}
}

type directoryReader struct{}

func (d *directoryReader) ReadFiles(ctx context.Context, paths []string) (<-chan File, error) {
	fileNames, err := d.getAllFileNames(paths)
	if err != nil {
		return nil, err
	}

	// TODO
	_ = fileNames

	return nil, err
}

func (d *directoryReader) getAllFileNames(paths []string) ([]string, error) {
	var fileNames []string

	for _, p := range paths {
		fileInfo, err := os.Stat(p)
		if err != nil {
			return nil, createErrFailedPath(err, p)
		}

		if !fileInfo.IsDir() {
			fileNames = append(fileNames, p)

			continue
		}

		files, err := d.readFilesInDir(p)
		if err != nil {
			return nil, err
		}

		fileNames = append(fileNames, files...)
	}

	return fileNames, nil
}

// readFilesInDir reads all the files from the dir or recursively reads all children directories and their files
func (d *directoryReader) readFilesInDir(dirPath string) ([]string, error) {
	dir, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, createErrFailedPath(err, dirPath)
	}

	var fileNames []string

	for _, entry := range dir {
		currPath := filepath.Join(dirPath, entry.Name())

		if !entry.IsDir() {
			fileNames = append(fileNames, currPath)
			continue
		}

		pths, err := d.readFilesInDir(currPath)
		if err != nil {
			return nil, err
		}
		fileNames = append(fileNames, pths...)
	}

	return fileNames, nil
}

func createErrFailedPath(err error, path string) error {
	return fmt.Errorf("failed to process path %s: %w", path, err)
}
