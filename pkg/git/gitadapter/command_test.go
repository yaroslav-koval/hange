package gitadapter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	commandexecutor_mock "github.com/yaroslav-koval/hange/mocks/commandexecutor"
)

func TestGitChangesCommandExecutor(t *testing.T) {
	t.Parallel()

	t.Run("argument substitution", func(t *testing.T) {
		t.Parallel()

		cem := commandexecutor_mock.NewMockCommandExecutor(t)

		ce := &gitChangesProvider{
			commandExecutor: cem,
		}

		cem.EXPECT().Output(mock.Anything, "git", []string{
			"--no-pager",
			"diff",
			"--staged",
			"--no-color",
			"--no-ext-diff",
			"--patch",
			"--unified=100",
		}).Return("resp", nil)

		res, err := ce.StagedDiff(t.Context(), 100)
		require.NoError(t, err)
		assert.Equal(t, "resp", res)
	})

	t.Run("git commit message", func(t *testing.T) {
		t.Parallel()

		cem := commandexecutor_mock.NewMockCommandExecutor(t)

		ce := &gitChangesProvider{
			commandExecutor: cem,
		}

		cem.EXPECT().Run(mock.Anything, "git", []string{
			"--no-pager",
			"commit",
			"-m",
			"user provided message",
		}).Return(nil)

		err := ce.Commit(t.Context(), "user provided message")
		require.NoError(t, err)
	})
}

func TestOsExecutor(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		scriptPath := filepath.Join(tempDir, "echo.sh")

		err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho success\n"), 0o500)
		require.NoError(t, err)

		executor := &osExecutor{}

		output, err := executor.Output(t.Context(), scriptPath)
		require.NoError(t, err)
		require.Equal(t, "success\n", output)
	})

	t.Run("success with args", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		scriptPath := filepath.Join(tempDir, "echo_args.sh")

		err := os.WriteFile(scriptPath, []byte("#!/bin/sh\necho \"arguments: $@\"\n"), 0o500)
		require.NoError(t, err)

		executor := &osExecutor{}

		output, err := executor.Output(t.Context(), scriptPath, "first", "second")
		require.NoError(t, err)
		require.Equal(t, "arguments: first second\n", output)
	})

	t.Run("failed", func(t *testing.T) {
		t.Parallel()

		tempDir := t.TempDir()
		scriptPath := filepath.Join(tempDir, "fail.sh")

		err := os.WriteFile(scriptPath, []byte("#!/bin/sh\nexit 1\n"), 0o500)
		require.NoError(t, err)

		executor := &osExecutor{}

		_, err = executor.Output(t.Context(), scriptPath)
		require.Error(t, err)
	})
}
