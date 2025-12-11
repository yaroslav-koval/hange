package openaiagent

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	explainprocessor_mock "github.com/yaroslav-koval/hange/mocks/explainprocessor"
	"github.com/yaroslav-koval/hange/pkg/agent"
)

func TestOpenAIUseCaseExplainFilesSuccess(t *testing.T) {
	ep := explainprocessor_mock.NewMockExplainProcessor(t)

	files := make(chan agent.File)
	close(files)

	ep.On("UploadFiles", mock.Anything, mock.Anything).Return(nil)
	ep.On("ExecuteExplainRequest", mock.Anything).Return("ok", nil)
	ep.On("Cleanup", mock.Anything)

	uc := &openAIUseCase{ep: ep}

	result, err := uc.ExplainFiles(context.Background(), files)
	require.NoError(t, err)
	require.Equal(t, "ok", result)
}

func TestOpenAIUseCaseExplainFilesUploadFails(t *testing.T) {
	ep := explainprocessor_mock.NewMockExplainProcessor(t)

	files := make(chan agent.File)
	close(files)

	uploadErr := errors.New("upload failed")

	ep.On("UploadFiles", mock.Anything, mock.Anything).Return(uploadErr)
	ep.On("Cleanup", mock.Anything)

	uc := &openAIUseCase{ep: ep}

	result, err := uc.ExplainFiles(context.Background(), files)
	require.ErrorIs(t, err, uploadErr)
	require.Empty(t, result)
	ep.AssertNotCalled(t, "ExecuteExplainRequest", mock.Anything)
}
