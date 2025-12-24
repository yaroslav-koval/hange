package errmapper

import (
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaroslav-koval/hange/pkg/fileprovider"
)

func TestOSFileErrMapper_Map(t *testing.T) {
	t.Parallel()

	t.Run("returns mapped error", func(t *testing.T) {
		t.Parallel()

		mapper := NewOSFileErrMapper()

		err := mapper.Map(os.ErrNotExist)
		require.Error(t, err)
		assert.Equal(t, fileprovider.ErrNotExist, err)
		assert.NotEqual(t, os.ErrNotExist, err)
	})

	t.Run("returns original error when no mapping", func(t *testing.T) {
		t.Parallel()

		mapper := NewOSFileErrMapper()
		expected := errors.New("some err")

		err := mapper.Map(expected)
		require.Error(t, err)
		assert.Equal(t, expected, err)
	})
}
