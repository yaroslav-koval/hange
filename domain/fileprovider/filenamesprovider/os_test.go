package filenamesprovider

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/yaroslav-koval/hange/domain/fileprovider"
	fileerrormapper_mock "github.com/yaroslav-koval/hange/mocks/fileerrormapper"
)

func TestOSFileNamesProvider_GetAllFileNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		layout  func(root string) []string
		wantRel []string
	}{
		{
			name: "file and nested dir with file and file",
			layout: func(root string) []string {
				paths := []string{
					filepath.Join(root, "a.txt"),
					filepath.Join(root, "dir", "b.txt"),
					filepath.Join(root, "c.txt"),
				}
				writeFiles(t, paths)
				return []string{
					filepath.Join(root, "a.txt"),
					filepath.Join(root, "dir"),
					filepath.Join(root, "c.txt"),
				}
			},
			wantRel: []string{"a.txt", "dir/b.txt", "c.txt"},
		},
		{
			name: "single file",
			layout: func(root string) []string {
				p := filepath.Join(root, "a.txt")
				writeFiles(t, []string{p})
				return []string{p}
			},
			wantRel: []string{"a.txt"},
		},
		{
			name: "single directory",
			layout: func(root string) []string {
				dir := filepath.Join(root, "dir")
				require.NoError(t, os.MkdirAll(dir, 0o755))
				return []string{dir}
			},
			wantRel: []string{},
		},
		{
			name: "dir with two files",
			layout: func(root string) []string {
				paths := []string{
					filepath.Join(root, "dir", "a.txt"),
					filepath.Join(root, "dir", "b.txt"),
				}
				writeFiles(t, paths)
				return []string{filepath.Dir(paths[0])}
			},
			wantRel: []string{"dir/a.txt", "dir/b.txt"},
		},
		{
			name: "deep nested directory",
			layout: func(root string) []string {
				p := filepath.Join(root, "dir", "dir", "dir", "file.txt")
				writeFiles(t, []string{p})
				return []string{filepath.Join(root, "dir")}
			},
			wantRel: []string{"dir/dir/dir/file.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			input := tt.layout(root)

			provider := NewOSFileNamesProvider(fileerrormapper_mock.NewMockFileErrorMapper(t))
			got, err := provider.GetAllFileNames(t.Context(), input)
			require.NoError(t, err)

			for i := range got {
				got[i], _ = filepath.Rel(root, got[i])
			}
			if got == nil {
				got = []string{}
			}
			assert.ElementsMatch(t, tt.wantRel, got)
		})
	}
}

func TestOSFileNamesProvider_GetAllFileNamesFailures(t *testing.T) {
	t.Parallel()

	t.Run("missing path", func(t *testing.T) {
		t.Parallel()

		errMapper := fileerrormapper_mock.NewMockFileErrorMapper(t)
		errMapper.EXPECT().Map(mock.Anything).Return(fileprovider.ErrNotExist)

		provider := NewOSFileNamesProvider(errMapper)
		missing := filepath.Join(t.TempDir(), "missing.txt")

		names, err := provider.GetAllFileNames(t.Context(), []string{missing})
		require.Error(t, err)
		assert.Nil(t, names)
		assert.ErrorIs(t, err, fileprovider.ErrNotExist)
		assert.ErrorContains(t, err, missing)
	})

	t.Run("permission denied", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		nested := filepath.Join(root, "dir", "nested")

		require.NoError(t, os.MkdirAll(nested, 0o755))
		require.NoError(t, os.Chmod(nested, 0o000))
		defer os.Chmod(nested, 0o755)

		errMapper := fileerrormapper_mock.NewMockFileErrorMapper(t)
		errMapper.EXPECT().Map(mock.Anything).Return(fileprovider.ErrPermission)

		provider := NewOSFileNamesProvider(errMapper)

		names, err := provider.GetAllFileNames(t.Context(), []string{root})
		require.Error(t, err)
		assert.Nil(t, names)
		assert.ErrorContains(t, err, nested)
		assert.ErrorIs(t, err, fileprovider.ErrPermission)
	})
}

func TestOSFileNamesProvider_GetAllFileNamesCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	provider := NewOSFileNamesProvider(fileerrormapper_mock.NewMockFileErrorMapper(t))

	names, err := provider.GetAllFileNames(ctx, []string{filepath.Join(t.TempDir(), "unused.txt")})
	require.Error(t, err)
	assert.Nil(t, names)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestOSFileNamesProvider_readFilesInDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		layout  func(root string) string
		wantRel []string
	}{
		{
			name: "mixed files and subdir",
			layout: func(root string) string {
				dir := filepath.Join(root, "dir")
				paths := []string{
					filepath.Join(dir, "a.txt"),
					filepath.Join(dir, "nested", "b.txt"),
					filepath.Join(dir, "c.txt"),
				}
				writeFiles(t, paths)
				return dir
			},
			wantRel: []string{"dir/a.txt", "dir/c.txt", "dir/nested/b.txt"},
		},
		{
			name: "only files",
			layout: func(root string) string {
				dir := filepath.Join(root, "dir")
				paths := []string{
					filepath.Join(dir, "a.txt"),
				}
				writeFiles(t, paths)
				return dir
			},
			wantRel: []string{"dir/a.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			dir := tt.layout(root)

			provider := NewOSFileNamesProvider(fileerrormapper_mock.NewMockFileErrorMapper(t)).(*osFileNamesProvider)
			got, err := provider.readFilesInDir(t.Context(), dir)
			require.NoError(t, err)

			for i := range got {
				got[i], _ = filepath.Rel(root, got[i])
			}
			assert.ElementsMatch(t, tt.wantRel, got)
		})
	}
}

func TestOSFileNamesProvider_readFilesInDirFailures(t *testing.T) {
	t.Parallel()

	t.Run("returns error when child directory cannot be read", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		child := filepath.Join(root, "child")
		grandchild := filepath.Join(child, "grandchild")

		require.NoError(t, os.MkdirAll(grandchild, 0o755))
		require.NoError(t, os.Chmod(grandchild, 0o000))
		defer os.Chmod(grandchild, 0o755)

		errMapper := fileerrormapper_mock.NewMockFileErrorMapper(t)
		errMapper.EXPECT().Map(mock.Anything).Return(fileprovider.ErrPermission)

		provider := NewOSFileNamesProvider(errMapper).(*osFileNamesProvider)
		_, err := provider.readFilesInDir(t.Context(), root)
		require.Error(t, err)
		assert.ErrorContains(t, err, grandchild)
		assert.ErrorIs(t, err, fileprovider.ErrPermission)
	})

	t.Run("returns context error when cancelled", func(t *testing.T) {
		t.Parallel()

		dir := t.TempDir()
		writeFiles(t, []string{filepath.Join(dir, "file.txt")})

		provider := NewOSFileNamesProvider(fileerrormapper_mock.NewMockFileErrorMapper(t)).(*osFileNamesProvider)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := provider.readFilesInDir(ctx, dir)
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func writeFiles(t *testing.T, paths []string) {
	t.Helper()
	for _, p := range paths {
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte("x"), 0o644))
	}
}
