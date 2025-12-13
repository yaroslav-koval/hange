package directory

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync/atomic"

	"github.com/yaroslav-koval/hange/pkg/entities"
	"github.com/yaroslav-koval/hange/pkg/fileprovider"
	"golang.org/x/sync/errgroup"
)

func NewDirectoryFileProvider() fileprovider.FileProvider {
	return &directoryReader{}
}

type directoryReader struct{}

// ReadFiles reads files and directories recursively. Second argument accepts both file paths and directory paths.
func (d *directoryReader) ReadFiles(ctx context.Context, paths []string) (<-chan entities.File, <-chan error, error) {
	fileNames, err := d.getAllFileNames(paths)
	if err != nil {
		return nil, nil, err
	}

	readerWorkers := 3
	if len(fileNames) < 10 {
		readerWorkers = 1
	}

	fileCh := make(chan entities.File, len(fileNames)*2)

	doneCh := d.readAndSendFile(ctx, fileNames, readerWorkers, fileCh)

	return fileCh, doneCh, nil
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

func (d *directoryReader) readAndSendFile(
	ctx context.Context, filePaths []string, workersCount int, filesCh chan<- entities.File) <-chan error {
	eg, ctx := errgroup.WithContext(ctx)

	fnIndex := atomic.Int32{}
	fnIndex.Add(-1)

	for range workersCount {
		eg.Go(func() error {
			for {
				// Select statement (at the end) doesn't guarantee order of cases execution.
				// So, the logic still can produce values several iterations even if context is cancelled.
				// Condition "ctx.Err() != nil" helps to prevent heavy os file reads in that case.
				if ctx.Err() != nil {
					return ctx.Err()
				}

				i := fnIndex.Add(1)
				if i >= int32(len(filePaths)) {
					return nil
				}

				fBytes, err := os.ReadFile(filePaths[i])
				if err != nil {
					return err
				}

				f := entities.File{
					Path: filePaths[i],
					Data: fBytes,
				}

				select {
				case <-ctx.Done():
					return context.Canceled
				case filesCh <- f:
				}
			}
		})
	}

	doneCh := make(chan error, 1)

	go func() {
		doneCh <- eg.Wait()
		close(doneCh)
		close(filesCh)
	}()

	return doneCh
}
