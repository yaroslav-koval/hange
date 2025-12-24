package filecontentprovider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	fileerrormapper_mock "github.com/yaroslav-koval/hange/mocks/fileerrormapper"
	"github.com/yaroslav-koval/hange/pkg/fileprovider"
)

func TestOSFileContentProvider_ReadFile(t *testing.T) {
	t.Parallel()

	t.Run("reads existing file", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		p := filepath.Join(dir, "file.txt")
		payload := []byte("hello")

		require.NoError(t, os.WriteFile(p, payload, 0o600))

		errMapper := fileerrormapper_mock.NewMockFileErrorMapper(t)

		provider := NewOSFileContentProvider(errMapper)

		data, err := provider.GetFileContent(context.Background(), p)
		require.NoError(t, err)
		assert.Equal(t, payload, data)
	})

	t.Run("maps missing file error", func(t *testing.T) {
		t.Parallel()

		errMapper := fileerrormapper_mock.NewMockFileErrorMapper(t)
		errMapper.EXPECT().Map(mock.Anything).Return(fileprovider.ErrNotExist)

		provider := NewOSFileContentProvider(errMapper)
		missing := filepath.Join(t.TempDir(), "missing.txt")

		data, err := provider.GetFileContent(context.Background(), missing)
		require.Error(t, err)
		assert.Nil(t, data)
		assert.ErrorIs(t, err, fileprovider.ErrNotExist)
	})
}
