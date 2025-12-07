package cmd

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	auth_mock "github.com/yaroslav-koval/hange/mocks/auth"
	"github.com/yaroslav-koval/hange/pkg/auth"
	"github.com/yaroslav-koval/hange/pkg/factory"
)

func TestAuthCommandUsesArgumentToken(t *testing.T) {
	mockAuth := auth_mock.NewMockAuth(t)
	mockAuth.EXPECT().SaveToken("abc123").Return(nil)

	err := runAuthCommand(t, mockAuth, []string{"abc123"})
	require.NoError(t, err)
}

func TestAuthCommandReadsTokenFromStdin(t *testing.T) {
	t.Cleanup(setStdin(t, "stdin-token"))

	mockAuth := auth_mock.NewMockAuth(t)
	mockAuth.EXPECT().SaveToken("stdin-token").Return(nil)

	err := runAuthCommand(t, mockAuth, nil)
	require.NoError(t, err)
}

func TestAuthCommandFailsWhenTokenMissing(t *testing.T) {
	t.Cleanup(setStdin(t, ""))

	mockAuth := auth_mock.NewMockAuth(t)

	err := runAuthCommand(t, mockAuth, nil)
	require.ErrorContains(t, err, "failed to parse token argument")
}

func TestAuthCommandPropagatesSaveError(t *testing.T) {
	mockAuth := auth_mock.NewMockAuth(t)
	saveErr := errors.New("save failed")
	mockAuth.EXPECT().SaveToken("bad-token").Return(saveErr)

	err := runAuthCommand(t, mockAuth, []string{"bad-token"})
	require.ErrorIs(t, err, saveErr)
}

func TestReadTokenFromStdinErrors(t *testing.T) {
	t.Run("when read fails", func(t *testing.T) {
		// Use a closed file to trigger read error
		tmp := t.TempDir()
		f, err := os.CreateTemp(tmp, "stdin")
		require.NoError(t, err)
		name := f.Name()
		require.NoError(t, f.Close())

		r, err := os.Open(name)
		require.NoError(t, err)
		require.NoError(t, r.Close()) // close so next read fails

		orig := os.Stdin
		os.Stdin = r
		t.Cleanup(func() { os.Stdin = orig })

		token, err := readTokenFromStdin()
		require.ErrorContains(t, err, "failed to read token from stdin")
		require.Empty(t, token)
	})
}

func runAuthCommand(t *testing.T, authService auth.Auth, args []string) error {
	t.Helper()

	cmd := &cobra.Command{RunE: authCmd.RunE}
	cmd.SetContext(context.Background())
	cmd.SetContext(appToCtx(cmd, &factory.App{Auth: authService}))

	return cmd.RunE(cmd, args)
}

func setStdin(t *testing.T, content string) func() {
	t.Helper()

	tmpDir := t.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "stdin")
	require.NoError(t, err)
	if content != "" {
		_, err = tmpFile.WriteString(content)
		require.NoError(t, err)
	}
	require.NoError(t, tmpFile.Close())

	r, err := os.Open(tmpFile.Name())
	require.NoError(t, err)

	originalStdin := os.Stdin
	os.Stdin = r

	return func() {
		os.Stdin = originalStdin
		_ = r.Close()
	}
}
