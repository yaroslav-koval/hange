package fileprovider

import (
	"context"
	"sync/atomic"

	"github.com/yaroslav-koval/hange/pkg/entities"
	"golang.org/x/sync/errgroup"
)

func NewFileProvider(cfg Config, fcProvider FileContentProvider) FileProvider {
	return &fileProvider{
		config:              cfg,
		fileContentProvider: fcProvider,
	}
}

type fileProvider struct {
	config              Config
	fileContentProvider FileContentProvider
}

// ReadFiles reads files and directories recursively. Second argument accepts both file paths and directory paths.
func (d *fileProvider) ReadFiles(ctx context.Context, fileNames []string) (<-chan entities.File, <-chan error) {
	fileCh := make(chan entities.File, d.config.BufferSize)

	doneCh := d.produceFiles(ctx, fileNames, d.config.Workers, fileCh)

	return fileCh, doneCh
}

func (d *fileProvider) produceFiles(
	ctx context.Context, filePaths []string, workersCount int, filesCh chan<- entities.File) <-chan error {
	eg, ctx := errgroup.WithContext(ctx)

	fnIndex := atomic.Int32{}
	fnIndex.Add(-1) // to start indexation from 0 after first 'fnIndex.Add(a)'

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

				fBytes, err := d.fileContentProvider.GetFileContent(ctx, filePaths[i])
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
