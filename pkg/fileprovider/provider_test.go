package fileprovider

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	filecontentprovider_mock "github.com/yaroslav-koval/hange/mocks/filecontentprovider"
	filenamesprovider_mock "github.com/yaroslav-koval/hange/mocks/filenamesprovider"
	"github.com/yaroslav-koval/hange/pkg/entities"
)

func TestFileProvider_ReadFilesSuccess(t *testing.T) {
	t.Parallel()

	fnProvider := filenamesprovider_mock.NewMockFileNamesProvider(t)
	fcProvider := filecontentprovider_mock.NewMockFileContentProvider(t)

	names := []string{"a.txt", "b.txt"}
	content := map[string][]byte{
		"a.txt": []byte("a"),
		"b.txt": []byte("b"),
	}

	fnProvider.EXPECT().GetAllFileNames(mock.Anything, []string{"ignored"}).Return(names, nil)
	fcProvider.EXPECT().GetFileContent(mock.Anything, "a.txt").Return(content["a.txt"], nil).Once()
	fcProvider.EXPECT().GetFileContent(mock.Anything, "b.txt").Return(content["b.txt"], nil).Once()

	fp := NewFileProvider(fnProvider, fcProvider)

	filesCh, doneCh, err := fp.ReadFiles(context.Background(), []string{"ignored"})
	require.NoError(t, err)

	got := make(map[string][]byte)
	for f := range filesCh {
		got[f.Path] = f.Data
	}
	assert.Equal(t, content, got)

	doneErr, ok := <-doneCh
	assert.True(t, ok)
	assert.NoError(t, doneErr)
}

func TestFileProvider_ReadFiles_FileNamesError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("names failure")

	fnProvider := filenamesprovider_mock.NewMockFileNamesProvider(t)
	fnProvider.EXPECT().GetAllFileNames(mock.Anything, mock.Anything).Return(nil, expectedErr)

	fp := NewFileProvider(fnProvider, filecontentprovider_mock.NewMockFileContentProvider(t))

	filesCh, doneCh, err := fp.ReadFiles(context.Background(), []string{"ignored"})
	require.Error(t, err)
	assert.Nil(t, filesCh)
	assert.Nil(t, doneCh)
	assert.ErrorIs(t, err, expectedErr)
}

func TestFileProvider_ReadFiles_FileContentError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("content failure")
	filePath := "file.txt"

	fnProvider := filenamesprovider_mock.NewMockFileNamesProvider(t)
	fnProvider.EXPECT().GetAllFileNames(mock.Anything, mock.Anything).Return([]string{filePath}, nil)

	fcProvider := filecontentprovider_mock.NewMockFileContentProvider(t)
	fcProvider.EXPECT().GetFileContent(mock.Anything, filePath).Return(nil, expectedErr)

	fp := NewFileProvider(fnProvider, fcProvider)

	filesCh, doneCh, err := fp.ReadFiles(context.Background(), []string{"ignored"})
	require.NoError(t, err)

	var files []entities.File
	for f := range filesCh {
		files = append(files, f)
	}
	assert.Empty(t, files)

	doneErr := <-doneCh
	assert.ErrorIs(t, doneErr, expectedErr)
}

func TestFileProvider_produceFilesCancelledContext(t *testing.T) {
	t.Parallel()

	blocker := filecontentprovider_mock.NewMockFileContentProvider(t)
	started := make(chan struct{})
	blocker.EXPECT().
		GetFileContent(mock.Anything, "file.txt").
		RunAndReturn(func(ctx context.Context, _ string) ([]byte, error) {
			close(started)
			<-ctx.Done()
			return nil, ctx.Err()
		})

	fp := &fileProvider{
		fileContentProvider: blocker,
	}

	ctx, cancel := context.WithCancel(context.Background())
	filesCh := make(chan entities.File, 1)
	doneCh := fp.produceFiles(ctx, []string{"file.txt"}, 1, filesCh)

	<-started
	cancel()

	doneErr := <-doneCh
	assert.ErrorIs(t, doneErr, context.Canceled)

	_, ok := <-filesCh
	assert.False(t, ok)
}
