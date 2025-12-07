package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetBuildConfigReturnsCopy(t *testing.T) {
	t.Cleanup(func() { buildConfig = nil })

	SetBuildConfig([]byte{1, 2, 3})

	cfgCopy := getBuildConfig()
	require.Equal(t, []byte{1, 2, 3}, cfgCopy)

	cfgCopy[0] = 9

	cfgCopyAgain := getBuildConfig()
	require.Equal(t, []byte{1, 2, 3}, cfgCopyAgain)
	require.NotEqual(t, &cfgCopy[0], &cfgCopyAgain[0])
}

func TestSetBuildConfigStoresData(t *testing.T) {
	t.Cleanup(func() { buildConfig = nil })

	SetBuildConfig([]byte{4, 5, 6})
	cfgCopy := getBuildConfig()
	require.Equal(t, []byte{4, 5, 6}, cfgCopy)

	SetBuildConfig([]byte{7, 8, 9})
	require.Equal(t, []byte{7, 8, 9}, getBuildConfig())
}
