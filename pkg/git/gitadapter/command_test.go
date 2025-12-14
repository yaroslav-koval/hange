package gitadapter

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitChangesCommandExecutor(t *testing.T) {
	t.Parallel()

	t.Run("status success", func(t *testing.T) {
		t.Parallel()

		cem := newCommandExecutorMock(t)
		defer cem.AssertCalled()

		ce := &gitChangesProvider{
			commandExecutor: cem,
		}

		cem.Expect(statusCommand, "resp", nil)

		res, err := ce.Status(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "resp", res)
	})

	t.Run("stat success", func(t *testing.T) {
		t.Parallel()

		cem := newCommandExecutorMock(t)
		defer cem.AssertCalled()

		ce := &gitChangesProvider{
			commandExecutor: cem,
		}

		cem.Expect(stagedStatusCommand, "resp", nil)

		res, err := ce.StagedStatus(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "resp", res)
	})

	t.Run("staged diff success", func(t *testing.T) {
		t.Parallel()

		cem := newCommandExecutorMock(t)
		defer cem.AssertCalled()

		ce := &gitChangesProvider{
			commandExecutor: cem,
		}

		cem.Expect(fmt.Sprintf(stagedDiffCommand, 10), "resp", nil)

		res, err := ce.StagedDiff(t.Context(), 10)
		require.NoError(t, err)
		assert.Equal(t, "resp", res)
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

		output, err := executor.Execute(t.Context(), scriptPath)
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

		output, err := executor.Execute(t.Context(), scriptPath+" first second")
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

		_, err = executor.Execute(t.Context(), scriptPath)
		require.Error(t, err)
	})
}
