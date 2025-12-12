package directory

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaroslav-koval/hange/pkg/entities"
	"github.com/yaroslav-koval/hange/pkg/graceful"
)

func TestReadFilesSuccess(t *testing.T) {
	t.Parallel()

	td := t.TempDir()
	require.NoError(t, os.WriteFile(path.Join(td, "a.txt"), []byte("x"), 0o600))
	require.NoError(t, os.WriteFile(path.Join(td, "b.txt"), []byte("x"), 0o600))
	require.NoError(t, os.WriteFile(path.Join(td, "c.txt"), []byte("x"), 0o600))

	dr := NewDirectoryFileProvider()
	filesCh, doneCh, err := dr.ReadFiles(graceful.Shutdown(t.Context()), []string{td})
	require.NoError(t, err)

	var files []entities.File

	for f := range filesCh {
		files = append(files, f)
	}

	assert.Equal(t, 3, len(files))

	e, ok := <-doneCh
	assert.True(t, ok)
	assert.NoError(t, e)
}

func TestReadFilesFail(t *testing.T) {
	t.Parallel()

	td := t.TempDir()
	require.NoError(t, os.WriteFile(path.Join(td, "a.txt"), []byte("x"), 0o600))
	require.NoError(t, os.WriteFile(path.Join(td, "b.txt"), []byte("x"), 0o000)) // not rights to read

	dr := NewDirectoryFileProvider()
	_, doneCh, err := dr.ReadFiles(graceful.Shutdown(t.Context()), []string{td})
	require.NoError(t, err)

	e, ok := <-doneCh
	assert.True(t, ok)
	assert.ErrorIs(t, e, os.ErrPermission)
}

func TestReadFilesMissingPath(t *testing.T) {
	t.Parallel()

	dr := NewDirectoryFileProvider()
	filesCh, doneCh, err := dr.ReadFiles(t.Context(), []string{"missing.txt"})

	require.Error(t, err)
	assert.Nil(t, filesCh)
	assert.Nil(t, doneCh)
	assert.ErrorIs(t, err, os.ErrNotExist)
	assert.ErrorContains(t, err, "missing.txt")
}

func TestReadFilesCancelledContext(t *testing.T) {
	t.Parallel()

	td := t.TempDir()
	p := filepath.Join(td, "a.txt")
	writeFiles(t, []string{p})

	ctx, cancel := context.WithCancel(graceful.Shutdown(t.Context()))
	filesCh := make(chan entities.File)

	dr := &directoryReader{}
	doneCh := dr.readAndSendFile(ctx, []string{p}, 1, filesCh)

	cancel()

	var doneErr error
	select {
	case doneErr = <-doneCh:
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for done channel")
	}
	assert.ErrorIs(t, doneErr, context.Canceled)

	_, ok := <-filesCh
	assert.False(t, ok)
}

func TestDirectoryReader_getAllFileNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		layout      func(root string) []string
		input       []string
		wantRel     []string
		expectError bool
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

			dr := &directoryReader{}
			got, err := dr.getAllFileNames(input)
			if tt.expectError {
				require.Error(t, err)
				return
			}
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

func TestDirectoryReader_readFilesInDir(t *testing.T) {
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

			dr := &directoryReader{}
			got, err := dr.readFilesInDir(dir)
			require.NoError(t, err)

			for i := range got {
				got[i], _ = filepath.Rel(root, got[i])
			}
			assert.ElementsMatch(t, tt.wantRel, got)
		})
	}
}

func TestParsingFailures(t *testing.T) {
	t.Parallel()

	t.Run("getAllFileNames returns error on missing path", func(t *testing.T) {
		t.Parallel()

		dr := &directoryReader{}
		_, err := dr.getAllFileNames([]string{"missing.txt"})
		require.Error(t, err)
		assert.ErrorIs(t, err, os.ErrNotExist)
		assert.ErrorContains(t, err, "missing.txt")
	})

	t.Run("getAllFileNames returns error when nested directory cannot be read", func(t *testing.T) {
		t.Parallel()

		root := t.TempDir()
		nested := filepath.Join(root, "dir", "nested")

		require.NoError(t, os.MkdirAll(nested, 0o755))
		require.NoError(t, os.Chmod(nested, 0o000))
		defer os.Chmod(nested, 0o755)

		dr := &directoryReader{}
		_, err := dr.getAllFileNames([]string{root})
		require.Error(t, err)
		assert.ErrorContains(t, err, nested)
	})

	t.Run("readFilesInDir returns error when child directory cannot be read", func(t *testing.T) {
		// this test covers case when there's an error in child directory
		t.Parallel()

		root := t.TempDir()
		child := filepath.Join(root, "child")
		grandchild := filepath.Join(child, "grandchild")

		require.NoError(t, os.MkdirAll(grandchild, 0o755))
		require.NoError(t, os.Chmod(grandchild, 0o000))
		defer os.Chmod(grandchild, 0o755)

		dr := &directoryReader{}
		_, err := dr.readFilesInDir(root)
		require.Error(t, err)
		assert.ErrorContains(t, err, grandchild)
	})

}

func writeFiles(t *testing.T, paths []string) {
	t.Helper()
	for _, p := range paths {
		require.NoError(t, os.MkdirAll(filepath.Dir(p), 0o755))
		require.NoError(t, os.WriteFile(p, []byte("x"), 0o644))
	}
}
