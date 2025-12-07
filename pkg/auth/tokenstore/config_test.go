package tokenstore

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/yaroslav-koval/hange/pkg/config/consts"
)

func TestStore(t *testing.T) {
	t.Parallel()

	mockCfg := &fakeConfigurator{}

	storer := NewConfigTokenStorer(mockCfg)
	err := storer.Store("token")

	require.NoError(t, err)
	assert.Equal(t, consts.AuthTokenPath, mockCfg.wroteField)
	assert.Equal(t, "token", mockCfg.wroteValue)
}

func TestStoreReturnsError(t *testing.T) {
	t.Parallel()

	mockCfg := &fakeConfigurator{writeErr: errors.New("write error")}

	storer := NewConfigTokenStorer(mockCfg)
	err := storer.Store("token")

	require.EqualError(t, err, "write error")
	assert.Equal(t, consts.AuthTokenPath, mockCfg.wroteField)
	assert.Equal(t, "token", mockCfg.wroteValue)
}

type fakeConfigurator struct {
	wroteField string
	wroteValue any
	writeErr   error
}

func (f *fakeConfigurator) WriteField(field string, value any) error {
	f.wroteField = field
	f.wroteValue = value
	return f.writeErr
}

func (f *fakeConfigurator) ReadField(field string) any {
	return nil
}
