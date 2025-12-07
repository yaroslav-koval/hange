package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReadFieldFromBytes(t *testing.T) {
	t.Parallel()

	file := []byte(`
app:
  version: v1`)

	v, err := ReadFieldFromBytes(file, FileTypeYaml, "app.version")
	require.NoError(t, err)
	assert.EqualValues(t, "v1", v)
}
