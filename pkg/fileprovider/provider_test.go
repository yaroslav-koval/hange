package fileprovider

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	filecontentprovider_mock "github.com/yaroslav-koval/hange/mocks/filecontentprovider"
	"github.com/yaroslav-koval/hange/pkg/entities"
)

func TestFileProvider_ReadFilesSuccess(t *testing.T) {
	t.Parallel()

	fcProvider := filecontentprovider_mock.NewMockFileContentProvider(t)

	names := []string{"a.txt", "b.txt"}
	content := map[string][]byte{
		"a.txt": []byte("a"),
		"b.txt": []byte("b"),
	}

	fcProvider.EXPECT().GetFileContent(mock.Anything, "a.txt").Return(content["a.txt"], nil).Once()
	fcProvider.EXPECT().GetFileContent(mock.Anything, "b.txt").Return(content["b.txt"], nil).Once()

	fp := NewFileProvider(Config{Workers: 2, BufferSize: 4}, fcProvider)

	filesCh, doneCh := fp.ReadFiles(context.Background(), names)

	got := make(map[string][]byte)
	for f := range filesCh {
		got[f.Path] = f.Data
	}
	assert.Equal(t, content, got)

	doneErr, ok := <-doneCh
	assert.True(t, ok)
	assert.NoError(t, doneErr)
}

func TestFileProvider_ReadFiles_ContentError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("content failure")
	filePath := "file.txt"

	fcProvider := filecontentprovider_mock.NewMockFileContentProvider(t)
	fcProvider.EXPECT().GetFileContent(mock.Anything, filePath).Return(nil, expectedErr)

	fp := NewFileProvider(Config{Workers: 1, BufferSize: 2}, fcProvider)

	filesCh, doneCh := fp.ReadFiles(context.Background(), []string{filePath})

	var files []entities.File
	for f := range filesCh {
		files = append(files, f)
	}
	assert.Empty(t, files)

	doneErr := <-doneCh
	assert.ErrorIs(t, doneErr, expectedErr)
}

func TestFileProvider_ReadFilesCancelledContext(t *testing.T) {
	t.Parallel()

	fcProvider := filecontentprovider_mock.NewMockFileContentProvider(t)
	started := make(chan struct{}, 1)
	fcProvider.EXPECT().
		GetFileContent(mock.Anything, "file.txt").
		RunAndReturn(func(ctx context.Context, _ string) ([]byte, error) {
			close(started)
			<-ctx.Done()
			return nil, ctx.Err()
		})

	fp := NewFileProvider(Config{Workers: 1, BufferSize: 1}, fcProvider)

	ctx, cancel := context.WithCancel(context.Background())
	filesCh, doneCh := fp.ReadFiles(ctx, []string{"file.txt"})

	<-started
	cancel()

	var doneErr error
	select {
	case doneErr = <-doneCh:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for done channel")
	}

	assert.ErrorIs(t, doneErr, context.Canceled)

	_, ok := <-filesCh
	assert.False(t, ok, "files channel should be closed")
}

func TestFileProvider_ReadFilesEmpty(t *testing.T) {
	t.Parallel()

	fp := NewFileProvider(Config{Workers: 1, BufferSize: 1}, filecontentprovider_mock.NewMockFileContentProvider(t))

	filesCh, doneCh := fp.ReadFiles(context.Background(), []string{})

	_, ok := <-filesCh
	assert.False(t, ok, "files channel should close immediately for no paths")

	doneErr := <-doneCh
	assert.NoError(t, doneErr)
}
