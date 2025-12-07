package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersionCommandSuccess(t *testing.T) {
	t.Cleanup(func() { buildConfig = nil })
	SetBuildConfig([]byte("version: 1.2.3"))

	err := versionCmd.RunE(versionCmd, nil)
	require.NoError(t, err)
}

func TestVersionCommandReadError(t *testing.T) {
	t.Cleanup(func() { buildConfig = nil })
	SetBuildConfig([]byte("version: [1, 2"))

	err := versionCmd.RunE(versionCmd, nil)
	require.Error(t, err)
}

func TestVersionCommandNilVersionValue(t *testing.T) {
	t.Cleanup(func() { buildConfig = nil })
	SetBuildConfig([]byte("foo: bar"))

	err := versionCmd.RunE(versionCmd, nil)
	require.NoError(t, err)
}
