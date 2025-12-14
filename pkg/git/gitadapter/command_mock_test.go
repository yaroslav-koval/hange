package gitadapter

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type commandExecutorMock struct {
	t               *testing.T
	expectedCommand string
	output          string
	err             error
	called          bool
}

func newCommandExecutorMock(t *testing.T) *commandExecutorMock {
	t.Helper()

	return &commandExecutorMock{
		t: t,
	}
}

func (m *commandExecutorMock) Expect(command string, output string, err error) {
	m.t.Helper()
	m.expectedCommand = command
	m.output = output
	m.err = err
}

func (m *commandExecutorMock) Execute(ctx context.Context, command string) (string, error) {
	m.t.Helper()
	m.called = true
	require.Equal(m.t, m.expectedCommand, command)

	return m.output, m.err
}

func (m *commandExecutorMock) AssertCalled() {
	m.t.Helper()
	require.True(m.t, m.called, "expected Execute to be called")
}
